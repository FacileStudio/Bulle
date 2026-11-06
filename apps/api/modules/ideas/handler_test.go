package ideas_test

import (
	"context"
	"encoding/json"
	"io"
	"iter"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/modules/ideas"
	"github.com/FacileStudio/Bulle/core/sandbox"
	"github.com/FacileStudio/nacelle"
	"github.com/go-chi/chi/v5"
)

type backend struct{ block chan struct{} }

func (b *backend) Name() string                                                { return "fake" }
func (b *backend) Capabilities() nacelle.Capabilities                          { return nacelle.Capabilities{} }
func (b *backend) CountTokens(context.Context, nacelle.Request) (int64, error) { return 0, nil }

func (b *backend) Stream(ctx context.Context, _ nacelle.Request) iter.Seq2[nacelle.Event, error] {
	return func(yield func(nacelle.Event, error) bool) {
		if b.block != nil {
			select {
			case <-b.block:
			case <-ctx.Done():
				yield(nacelle.Event{}, ctx.Err())
				return
			}
		}
		if yield(nacelle.Event{Kind: nacelle.KindText, Text: "ok"}, nil) {
			yield(nacelle.Event{Kind: nacelle.KindDone, Stop: nacelle.StopEnd}, nil)
		}
	}
}

type api struct {
	t      *testing.T
	router chi.Router
	store  *store.Memory
}

func newAPI(t *testing.T, b nacelle.Backend) *api {
	t.Helper()
	st := store.NewMemory()
	rn := runner.New(runner.Config{
		Store:     st,
		Backends:  func(string, string) (nacelle.Backend, error) { return b, nil },
		VMTools:   func(*sandbox.Target) ([]nacelle.Tool, io.Closer, error) { return nil, io.NopCloser(nil), nil },
		HostTools: func(int64) ([]nacelle.Tool, error) { return nil, nil },
	})
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		rn.Shutdown(ctx)
	})
	r := chi.NewRouter()
	requireAuth := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			id, _ := strconv.ParseInt(req.Header.Get("X-User"), 10, 64)
			identity := authcontext.Identity{UserID: id, Email: "user@facile.studio"}
			next.ServeHTTP(w, req.WithContext(authcontext.WithIdentity(req.Context(), identity)))
		})
	}
	ideas.RegisterRoutes(r, st, rn, nil, requireAuth)
	return &api{t: t, router: r, store: st}
}

func (a *api) call(user, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("X-User", user)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	a.router.ServeHTTP(rec, req)
	return rec
}

func (a *api) space(user int64) store.Space {
	space, err := a.store.EnsureDefaultSpace(context.Background(), user)
	if err != nil {
		a.t.Fatal(err)
	}
	return space
}

func (a *api) idea(user int64, spaceID, draft string) store.Idea {
	idea, err := a.store.CreateIdea(context.Background(), store.Idea{
		UserID: user, SpaceID: spaceID, BatchID: "b", Kind: "mail", Title: "Reply to Ann",
		Why: "She asked", Prompt: "Reply to Ann about the quote", Draft: draft,
		Fingerprint: "courrier:5/email:812/reply", ExpiresAt: time.Now().Add(time.Hour),
	})
	if err != nil {
		a.t.Fatal(err)
	}
	return idea
}

func (a *api) waitIdle(user int64, spaceID string) {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if _, running, _ := a.store.RunningIdeaBatch(context.Background(), user, spaceID); !running {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	a.t.Fatal("the batch never finished")
}

func TestListCreatesTheSettingsRowWithDefaults(t *testing.T) {
	a := newAPI(t, &backend{})
	space := a.space(1)
	rec := a.call("1", http.MethodGet, "/ideas?space_id="+space.ID, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /ideas = %d %s", rec.Code, rec.Body)
	}
	var page ideas.ListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	want := ideas.SettingsResponse{SpaceID: space.ID, Enabled: true, Hour: 8, Timezone: "UTC"}
	if page.Settings != want || page.Ideas == nil || page.Goals == nil || page.Generating || page.LastBatchAt != nil {
		t.Fatalf("page = %+v", page)
	}
	if _, err := a.store.GetIdeaSettings(context.Background(), 1, space.ID); err != nil {
		t.Fatalf("GET /ideas did not create the settings row: %v", err)
	}
	if rec := a.call("2", http.MethodGet, "/ideas?space_id="+space.ID, ""); rec.Code != http.StatusNotFound {
		t.Fatalf("another user's space = %d", rec.Code)
	}
}

func TestAnotherUsersIdeaIsNotFound(t *testing.T) {
	a := newAPI(t, &backend{})
	idea := a.idea(1, a.space(1).ID, "")
	a.space(2)
	if rec := a.call("2", http.MethodPatch, "/ideas/"+idea.ID, `{"status":"saved"}`); rec.Code != http.StatusNotFound {
		t.Fatalf("PATCH another user's idea = %d", rec.Code)
	}
	if rec := a.call("2", http.MethodPost, "/ideas/"+idea.ID+"/start", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("start another user's idea = %d", rec.Code)
	}
	rec := a.call("1", http.MethodPatch, "/ideas/"+idea.ID, `{"status":"saved","feedback":"up"}`)
	var got ideas.IdeaResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	if rec.Code != http.StatusOK || got.Status != store.IdeaSaved || got.Feedback != store.FeedbackUp {
		t.Fatalf("PATCH own idea = %d %s", rec.Code, rec.Body)
	}
	if rec := a.call("1", http.MethodPatch, "/ideas/"+idea.ID, `{"status":"done"}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("PATCH to done = %d", rec.Code)
	}
}

func TestStartOpensAWebChatInTheIdeasSpace(t *testing.T) {
	a := newAPI(t, &backend{})
	ctx := context.Background()
	a.space(1)
	work, _ := a.store.CreateSpace(ctx, store.Space{UserID: 1, Name: "Work"})
	idea := a.idea(1, work.ID, "Hi Ann,")
	rec := a.call("1", http.MethodPost, "/ideas/"+idea.ID+"/start", "")
	var res ideas.StartResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &res)
	if rec.Code != http.StatusOK || res.SessionID == "" {
		t.Fatalf("start = %d %s", rec.Code, rec.Body)
	}
	session, err := a.store.GetSession(ctx, 1, res.SessionID)
	if err != nil || session.SpaceID != work.ID || session.Channel != store.ChannelWeb || session.Title != idea.Title {
		t.Fatalf("session = %+v %v", session, err)
	}
	messages, _ := a.store.Messages(ctx, res.SessionID)
	if len(messages) == 0 || !strings.Contains(string(messages[0].Payload), `Reply to Ann about the quote\n\nDraft:\nHi Ann,`) {
		t.Fatalf("the first message is not the prompt and draft: %v", messages)
	}
	got, _ := a.store.GetIdea(ctx, 1, idea.ID)
	if got.Status != store.IdeaDone || got.SessionID != res.SessionID {
		t.Fatalf("idea after start = %+v", got)
	}
	again := a.call("1", http.MethodPost, "/ideas/"+idea.ID+"/start", "")
	if !strings.Contains(again.Body.String(), res.SessionID) {
		t.Fatalf("starting a done idea again opened another chat: %s", again.Body)
	}
}

func TestManualGenerateIsBusyThenLimited(t *testing.T) {
	block := make(chan struct{})
	a := newAPI(t, &backend{block: block})
	space := a.space(1)
	body := `{"space_id":"` + space.ID + `"}`
	if rec := a.call("1", http.MethodPost, "/ideas/generate", body); rec.Code != http.StatusAccepted {
		t.Fatalf("first generate = %d %s", rec.Code, rec.Body)
	}
	if rec := a.call("1", http.MethodPost, "/ideas/generate", body); rec.Code != http.StatusConflict {
		t.Fatalf("generate while running = %d", rec.Code)
	}
	close(block)
	a.waitIdle(1, space.ID)
	for i := 2; i <= ideas.ManualBatchesPerDay; i++ {
		if rec := a.call("1", http.MethodPost, "/ideas/generate", body); rec.Code != http.StatusAccepted {
			t.Fatalf("generate %d = %d %s", i, rec.Code, rec.Body)
		}
		a.waitIdle(1, space.ID)
	}
	if rec := a.call("1", http.MethodPost, "/ideas/generate", body); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("generate past the daily cap = %d %s", rec.Code, rec.Body)
	}
}

func TestSettingsAndGoals(t *testing.T) {
	a := newAPI(t, &backend{})
	space := a.space(1)
	a.space(2)
	rec := a.call("1", http.MethodPut, "/ideas/settings", `{"space_id":"`+space.ID+`","timezone":"Europe/Paris","hour":7,"push":true}`)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"timezone":"Europe/Paris"`) {
		t.Fatalf("PUT settings = %d %s", rec.Code, rec.Body)
	}
	if rec := a.call("1", http.MethodPut, "/ideas/settings", `{"space_id":"`+space.ID+`","timezone":"Mars/Olympus"}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("a bad timezone = %d", rec.Code)
	}
	rec = a.call("1", http.MethodPost, "/ideas/goals", `{"space_id":"`+space.ID+`","title":"Run 10k","category":"health"}`)
	var goal ideas.GoalResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &goal)
	if rec.Code != http.StatusCreated || goal.ID == "" {
		t.Fatalf("POST goal = %d %s", rec.Code, rec.Body)
	}
	if rec := a.call("1", http.MethodPost, "/ideas/goals", `{"space_id":"`+space.ID+`","title":"x","category":"fame"}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("a bad category = %d", rec.Code)
	}
	if rec := a.call("2", http.MethodPut, "/ideas/goals/"+goal.ID, `{"done":true}`); rec.Code != http.StatusNotFound {
		t.Fatalf("PUT another user's goal = %d", rec.Code)
	}
	if rec := a.call("1", http.MethodPut, "/ideas/goals/"+goal.ID, `{"done":true}`); !strings.Contains(rec.Body.String(), `"done":true`) {
		t.Fatalf("PUT goal = %d %s", rec.Code, rec.Body)
	}
	if rec := a.call("2", http.MethodDelete, "/ideas/goals/"+goal.ID, ""); rec.Code != http.StatusNotFound {
		t.Fatalf("DELETE another user's goal = %d", rec.Code)
	}
	if rec := a.call("1", http.MethodDelete, "/ideas/goals/"+goal.ID, ""); rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE goal = %d", rec.Code)
	}
}
