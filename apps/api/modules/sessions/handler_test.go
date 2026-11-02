package sessions

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"iter"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/core/sandbox"
	"github.com/FacileStudio/nacelle"
	"github.com/go-chi/chi/v5"
)

type fakeComputerProvider struct {
	touched atomic.Int64
}

func (f *fakeComputerProvider) Ensure(_ context.Context, _ string) (*sandbox.Target, error) {
	return &sandbox.Target{Name: "test-box", Status: "running", Backend: "boite", User: "boite"}, nil
}

func (f *fakeComputerProvider) Touch(string) { f.touched.Add(1) }

type fakeBackend struct{}

func (f *fakeBackend) Name() string { return "fake" }

func (f *fakeBackend) Capabilities() nacelle.Capabilities {
	return nacelle.Capabilities{Thinking: true}
}

func (f *fakeBackend) CountTokens(context.Context, nacelle.Request) (int64, error) { return 0, nil }

func (f *fakeBackend) Stream(_ context.Context, _ nacelle.Request) iter.Seq2[nacelle.Event, error] {
	return func(yield func(nacelle.Event, error) bool) {
		if !yield(nacelle.Event{Kind: nacelle.KindText, Text: "Hello there"}, nil) {
			return
		}
		yield(nacelle.Event{Kind: nacelle.KindDone, Stop: nacelle.StopEnd, Usage: nacelle.Usage{InputTokens: 10, OutputTokens: 5, Cost: 0.001}}, nil)
	}
}

func setupTestRouter(t *testing.T) (chi.Router, store.Store, *runner.Runner) {
	t.Helper()
	st := store.NewMemory()
	rn := runner.New(runner.Config{
		Store:     st,
		Computers: &fakeComputerProvider{},
		Backends: func(b, m string) (nacelle.Backend, error) {
			return &fakeBackend{}, nil
		},
		VMTools: func(target *sandbox.Target) ([]nacelle.Tool, io.Closer, error) {
			return nil, io.NopCloser(nil), nil
		},
		HostTools: func() ([]nacelle.Tool, error) {
			return nil, nil
		},
	})
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		rn.Shutdown(ctx)
	})

	r := chi.NewRouter()
	requireAuth := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			identity := authcontext.Identity{UserID: 1, Email: "user@example.com"}
			next.ServeHTTP(w, req.WithContext(authcontext.WithIdentity(req.Context(), identity)))
		})
	}
	RegisterRoutes(r, st, rn, nil, requireAuth)
	return r, st, rn
}

func TestListAndCreateSessions(t *testing.T) {
	router, st, _ := setupTestRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sessions", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if strings.TrimSpace(w.Body.String()) != "[]" {
		t.Fatalf("expected empty array, got %s", w.Body.String())
	}

	body := bytes.NewBufferString(`{"title":"First Session"}`)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/sessions", body)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var created CreateSessionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.ID == "" || created.Title != "First Session" {
		t.Fatalf("unexpected created response: %+v", created)
	}

	items, err := st.ListSessions(context.Background(), 1)
	if err != nil || len(items) != 1 {
		t.Fatalf("expected 1 session in store, got %d (err: %v)", len(items), err)
	}
}

func TestGetAndDeleteSession(t *testing.T) {
	router, st, _ := setupTestRouter(t)

	session, err := st.CreateSession(context.Background(), store.Session{
		ID:        "s1",
		UserID:    1,
		Title:     "Test Session",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}

	rawUser, _ := runner.EncodeMessage(nacelle.UserText("hello"), nil)
	rawAssistant, _ := runner.EncodeMessage(nacelle.AssistantText("hi"), nil)
	_ = st.ReplaceMessages(context.Background(), session.ID, []json.RawMessage{rawUser, rawAssistant})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sessions/s1", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var detail SessionDetailResponse
	if err := json.Unmarshal(w.Body.Bytes(), &detail); err != nil {
		t.Fatal(err)
	}
	if detail.Session.ID != "s1" || len(detail.Messages) != 2 {
		t.Fatalf("unexpected detail: %+v", detail)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/sessions/s1", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/sessions/s1", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestGetSessionSplitsAUserTurnTheEngineMerged(t *testing.T) {
	router, st, _ := setupTestRouter(t)

	session, err := st.CreateSession(context.Background(), store.Session{
		ID: "s-merged", UserID: 1, Title: "Merged", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	merged, _ := runner.EncodeMessage(nacelle.Message{
		Role:  nacelle.RoleUser,
		Parts: []nacelle.Part{nacelle.Text{Text: "first"}, nacelle.Text{Text: "second"}},
	}, nil)
	answer, _ := runner.EncodeMessage(nacelle.Message{
		Role:  nacelle.RoleAssistant,
		Parts: []nacelle.Part{nacelle.Text{Text: "one"}, nacelle.Text{Text: "two"}},
	}, nil)
	_ = st.ReplaceMessages(context.Background(), session.ID, []json.RawMessage{merged, answer})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/sessions/s-merged", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var detail SessionDetailResponse
	if err := json.Unmarshal(w.Body.Bytes(), &detail); err != nil {
		t.Fatal(err)
	}
	want := []StoredMessage{
		{Seq: 1, Role: "user", Content: "first"},
		{Seq: 2, Role: "user", Content: "second"},
		{Seq: 3, Role: "assistant", Content: "one\n\ntwo"},
	}
	if len(detail.Messages) != len(want) {
		t.Fatalf("messages = %+v", detail.Messages)
	}
	for i := range want {
		got := detail.Messages[i]
		if got.Seq != want[i].Seq || got.Role != want[i].Role || got.Content != want[i].Content || got.Thinking != want[i].Thinking || len(got.Tools) != len(want[i].Tools) {
			t.Fatalf("message %d = %+v, want %+v", i, detail.Messages[i], want[i])
		}
	}
}

func TestGetSessionReturnsErroredRun(t *testing.T) {
	router, st, _ := setupTestRouter(t)

	session, err := st.CreateSession(context.Background(), store.Session{
		ID:        "s-err",
		UserID:    1,
		Title:     "Errored Run Session",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}

	rawUser, _ := runner.EncodeMessage(nacelle.UserText("help"), nil)
	_ = st.ReplaceMessages(context.Background(), session.ID, []json.RawMessage{rawUser})

	run := store.Run{ID: "run-err-1", SessionID: session.ID, Status: store.RunRunning, StartedAt: time.Now()}
	if err := st.CreateRun(context.Background(), run); err != nil {
		t.Fatal(err)
	}
	if err := st.FinishRun(context.Background(), run.ID, store.RunError, "error", nil, 0); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sessions/s-err", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var detail SessionDetailResponse
	if err := json.Unmarshal(w.Body.Bytes(), &detail); err != nil {
		t.Fatal(err)
	}
	if detail.ActiveRunID != "run-err-1" {
		t.Fatalf("expected active_run_id run-err-1, got %q", detail.ActiveRunID)
	}
}

func TestSendMessageAndStop(t *testing.T) {
	router, st, _ := setupTestRouter(t)

	_, err := st.CreateSession(context.Background(), store.Session{
		ID:        "s2",
		UserID:    1,
		Title:     "Chat Session",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}

	body := bytes.NewBufferString(`{"text":"hello agent"}`)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/sessions/s2/messages", body)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", w.Code, w.Body.String())
	}
	var res SendMessageResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.RunID == "" {
		t.Fatal("expected run_id")
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/runs/"+res.RunID+"/stop", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestStreamEventsReplay(t *testing.T) {
	router, st, _ := setupTestRouter(t)

	_, err := st.CreateSession(context.Background(), store.Session{
		ID:        "s3",
		UserID:    1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	run := store.Run{ID: "r3", SessionID: "s3", Status: store.RunFinished}
	if err := st.CreateRun(context.Background(), run); err != nil {
		t.Fatal(err)
	}
	_ = st.AppendEvent(context.Background(), store.Event{RunID: "r3", Seq: 1, Type: runner.EventRunStarted, Payload: []byte(`{"run_id":"r3"}`)})
	_ = st.AppendEvent(context.Background(), store.Event{RunID: "r3", Seq: 2, Type: runner.EventText, Payload: []byte(`{"delta":"hello"}`)})
	_ = st.AppendEvent(context.Background(), store.Event{RunID: "r3", Seq: 3, Type: runner.EventRunFinished, Payload: []byte(`{"stop":"end"}`)})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/runs/r3/events?last_event_id=1", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "event: TEXT_MESSAGE_CONTENT") {
		t.Fatalf("expected text event in replay, got: %s", body)
	}
	if !strings.Contains(body, "event: RUN_FINISHED") {
		t.Fatalf("expected finished event in replay, got: %s", body)
	}
	if strings.Contains(body, "event: RUN_STARTED") {
		t.Fatalf("event with seq 1 should have been filtered by last_event_id=1")
	}

	scanner := bufio.NewScanner(strings.NewReader(body))
	foundSeq2 := false
	for scanner.Scan() {
		line := scanner.Text()
		if line == "id: 2" {
			foundSeq2 = true
		}
	}
	if !foundSeq2 {
		t.Fatalf("expected 'id: 2' in output, got: %s", body)
	}
}

func TestGetSessionRestoresThinkingAndTools(t *testing.T) {
	router, st, _ := setupTestRouter(t)

	session, err := st.CreateSession(context.Background(), store.Session{
		ID:        "s-tools",
		UserID:    1,
		Title:     "Tools Session",
		Channel:   store.ChannelCron,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}

	userMsg, _ := runner.EncodeMessage(nacelle.UserText("run command"), nil)
	asstMsg, _ := runner.EncodeMessage(nacelle.Message{
		Role: nacelle.RoleAssistant,
		Parts: []nacelle.Part{
			nacelle.Reasoning{Text: "deciding to call tool"},
			nacelle.Text{Text: "Let me check."},
			nacelle.ToolCall{
				ID:    "c1",
				Name:  "bash",
				Input: json.RawMessage(`{"command":"echo hello"}`),
			},
		},
	}, nil)
	toolRes, _ := runner.EncodeMessage(nacelle.Message{
		Role: nacelle.RoleUser,
		Parts: []nacelle.Part{
			nacelle.ToolResult{
				ID:     "c1",
				Name:   "bash",
				Result: "hello\n",
				Failed: false,
			},
		},
	}, nil)
	asstAnswer, _ := runner.EncodeMessage(nacelle.Message{
		Role: nacelle.RoleAssistant,
		Parts: []nacelle.Part{
			nacelle.Text{Text: "all done"},
		},
	}, nil)

	_ = st.ReplaceMessages(context.Background(), session.ID, []json.RawMessage{userMsg, asstMsg, toolRes, asstAnswer})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/sessions/s-tools", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var detail SessionDetailResponse
	if err := json.Unmarshal(w.Body.Bytes(), &detail); err != nil {
		t.Fatal(err)
	}
	if len(detail.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d: %+v", len(detail.Messages), detail.Messages)
	}

	userGot := detail.Messages[0]
	if userGot.Seq != 1 || userGot.Role != "user" || userGot.Content != "run command" {
		t.Fatalf("unexpected user message: %+v", userGot)
	}

	asstGot := detail.Messages[1]
	if asstGot.Seq != 2 || asstGot.Role != "assistant" || asstGot.Content != "Let me check.\n\nall done" {
		t.Fatalf("unexpected assistant message: %+v", asstGot)
	}
	if asstGot.Answer != "all done" || detail.Session.Channel != store.ChannelCron {
		t.Fatalf("answer %q channel %q, want the final text and cron", asstGot.Answer, detail.Session.Channel)
	}
	if asstGot.Thinking != "deciding to call tool" {
		t.Fatalf("unexpected thinking: %q", asstGot.Thinking)
	}
	if len(asstGot.Tools) != 1 {
		t.Fatalf("expected 1 tool call, got %d: %+v", len(asstGot.Tools), asstGot.Tools)
	}
	toolGot := asstGot.Tools[0]
	if toolGot.CallID != "c1" || toolGot.Name != "bash" || toolGot.Status != "ok" || toolGot.OutputPreview != "hello\n" {
		t.Fatalf("unexpected tool call: %+v", toolGot)
	}
	argsMap, ok := toolGot.Args.(map[string]any)
	if !ok || argsMap["command"] != "echo hello" {
		t.Fatalf("unexpected tool args: %+v", toolGot.Args)
	}
}

func TestGetSessionRestoresFailedToolCall(t *testing.T) {
	router, st, _ := setupTestRouter(t)

	session, err := st.CreateSession(context.Background(), store.Session{
		ID:        "s-failed-tool",
		UserID:    1,
		Title:     "Failed Tool Session",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}

	userMsg, _ := runner.EncodeMessage(nacelle.UserText("run failing command"), nil)
	asstMsg, _ := runner.EncodeMessage(nacelle.Message{
		Role: nacelle.RoleAssistant,
		Parts: []nacelle.Part{
			nacelle.ToolCall{
				ID:    "c-fail",
				Name:  "bash",
				Input: json.RawMessage(`{"command":"exit 1"}`),
			},
		},
	}, nil)
	toolRes, _ := runner.EncodeMessage(nacelle.Message{
		Role: nacelle.RoleUser,
		Parts: []nacelle.Part{
			nacelle.ToolResult{
				ID:     "c-fail",
				Name:   "bash",
				Result: "command returned error code 1",
				Failed: true,
			},
		},
	}, nil)

	_ = st.ReplaceMessages(context.Background(), session.ID, []json.RawMessage{userMsg, asstMsg, toolRes})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/sessions/s-failed-tool", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var detail SessionDetailResponse
	if err := json.Unmarshal(w.Body.Bytes(), &detail); err != nil {
		t.Fatal(err)
	}
	if len(detail.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d: %+v", len(detail.Messages), detail.Messages)
	}
	asstGot := detail.Messages[1]
	if len(asstGot.Tools) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(asstGot.Tools))
	}
	toolGot := asstGot.Tools[0]
	if toolGot.Status != "failed" || toolGot.OutputPreview != "command returned error code 1" {
		t.Fatalf("expected failed status and output preview, got %+v", toolGot)
	}
}

func TestGetSessionOmitsTrailingAssistantWithActiveRun(t *testing.T) {
	router, st, _ := setupTestRouter(t)

	session, err := st.CreateSession(context.Background(), store.Session{
		ID:        "s-active",
		UserID:    1,
		Title:     "Active Session",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}

	userMsg, _ := runner.EncodeMessage(nacelle.UserText("hello"), nil)
	asstMsg, _ := runner.EncodeMessage(nacelle.Message{
		Role: nacelle.RoleAssistant,
		Parts: []nacelle.Part{
			nacelle.ToolCall{
				ID:    "c-active",
				Name:  "bash",
				Input: json.RawMessage(`{"command":"sleep 10"}`),
			},
		},
	}, nil)
	_ = st.ReplaceMessages(context.Background(), session.ID, []json.RawMessage{userMsg, asstMsg})

	run := store.Run{ID: "run-act-1", SessionID: session.ID, Status: store.RunRunning, StartedAt: time.Now()}
	if err := st.CreateRun(context.Background(), run); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/sessions/s-active", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var detail SessionDetailResponse
	if err := json.Unmarshal(w.Body.Bytes(), &detail); err != nil {
		t.Fatal(err)
	}
	if detail.ActiveRunID != "run-act-1" {
		t.Fatalf("expected active run id run-act-1, got %q", detail.ActiveRunID)
	}
	if len(detail.Messages) != 1 {
		t.Fatalf("expected 1 message (assistant omitted), got %d: %+v", len(detail.Messages), detail.Messages)
	}
	if detail.Messages[0].Role != "user" {
		t.Fatalf("expected user message, got %+v", detail.Messages[0])
	}
}
