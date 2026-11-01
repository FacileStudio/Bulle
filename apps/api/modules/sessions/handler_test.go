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
