package ideas_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"iter"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/middleware"
	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/modules/ideas"
	"github.com/FacileStudio/Bulle/apps/api/modules/sessions"
	"github.com/FacileStudio/Bulle/core/nacelletest"
	"github.com/FacileStudio/Bulle/core/sandbox"
	"github.com/FacileStudio/nacelle"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/session"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/go-chi/chi/v5"
)

const (
	gloutonRead  = "glouton.read.token"
	gloutonWrite = "glouton.write.token"
	otherScope   = "glouton.scope.token"
	noActor      = "bulle.noact.token"
)

type scripted struct{ turns []nacelletest.Turn }

func (s scripted) Name() string                                                { return "script" }
func (s scripted) Capabilities() nacelle.Capabilities                          { return nacelle.Capabilities{Thinking: true} }
func (s scripted) CountTokens(context.Context, nacelle.Request) (int64, error) { return 0, nil }

func (s scripted) Stream(ctx context.Context, req nacelle.Request) iter.Seq2[nacelle.Event, error] {
	return nacelle.NewBackend(&nacelletest.Script{Turns: s.turns}).Stream(ctx, req)
}

type tokens map[string]porte.Identity

func (t tokens) VerifyJWT(_ context.Context, raw string) (porte.Identity, error) {
	if identity, ok := t[raw]; ok {
		return identity, nil
	}
	return porte.Identity{}, errors.New("bad token")
}

type noSessions struct{ porte.SessionStore }

type resolver struct{}

func (resolver) IdentityForUser(_ context.Context, userID int64) (authcontext.Identity, error) {
	return authcontext.Identity{UserID: userID, Email: "user@facile.studio"}, nil
}

type plan struct{ paid bool }

func (p plan) CheckAIIn(context.Context, int64, string) error {
	if p.paid {
		return nil
	}
	return troncerrors.New("payment_required", "upgrade to use AI", nil).WithUpgradeURL("https://registre.facile.studio/spaces")
}

type appAPI struct {
	t      *testing.T
	router chi.Router
	store  *store.Memory
}

func newAppAPI(t *testing.T, backend nacelle.Backend, access runner.Access) *appAPI {
	t.Helper()
	st := store.NewMemory()
	create, err := nacelle.NewTool("opus_create_task", "Create a task", func(context.Context, struct {
		Title string `json:"title"`
	}) (string, error) {
		return `{"id":"t-42","url":"https://opus.facile.studio/tasks/t-42","title":"Call Acme"}`, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	rn := runner.New(runner.Config{
		Store:     st,
		Backends:  func(string, string) (nacelle.Backend, error) { return backend, nil },
		VMTools:   func(*sandbox.Target) ([]nacelle.Tool, io.Closer, error) { return nil, io.NopCloser(nil), nil },
		HostTools: func(int64) ([]nacelle.Tool, error) { return []nacelle.Tool{create}, nil },
		Outward:   []string{"opus_create_task"},
		Access:    access,
		Logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		rn.Shutdown(ctx)
	})
	manager, err := session.New(porte.Config{SuccessURL: "https://bulle.test/", ScopePrefix: "bulle"}, session.Deps{Sessions: noSessions{}})
	if err != nil {
		t.Fatal(err)
	}
	manager.WithJWT(tokens{
		gloutonRead:  {UserID: 1, Subject: "sub-1", Scopes: []string{"bulle.read"}, Actor: "glouton"},
		gloutonWrite: {UserID: 1, Subject: "sub-1", Scopes: []string{"bulle.read", "bulle.write"}, Actor: "glouton"},
		otherScope:   {UserID: 1, Subject: "sub-1", Scopes: []string{"glouton.read", "glouton.write"}, Actor: "glouton"},
		noActor:      {UserID: 1, Subject: "sub-1", Scopes: []string{"bulle.read", "bulle.write"}},
	})
	requireAuth := middleware.RequireAuth(manager, resolver{})
	r := chi.NewRouter()
	r.Route("/api", func(api chi.Router) {
		ideas.RegisterRoutes(api, st, rn, nil, requireAuth)
		sessions.RegisterRoutes(api, st, rn, nil, requireAuth)
	})
	return &appAPI{t: t, router: r, store: st}
}

func (a *appAPI) call(token, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	a.router.ServeHTTP(rec, req)
	return rec
}

func (a *appAPI) gloutonIdea(spaceID string) store.Idea {
	idea, err := a.store.CreateIdea(context.Background(), store.Idea{
		UserID: 1, SpaceID: spaceID, BatchID: "b", App: "glouton", Kind: "follow_up", Title: "Follow up Acme",
		Why: "Acme replied", Prompt: "Create a task to call Acme", Fingerprint: "glouton:company:7/follow-up",
		ExpiresAt: time.Now().Add(time.Hour),
	})
	if err != nil {
		a.t.Fatal(err)
	}
	return idea
}

func (a *appAPI) pollRun(runID string, done func(ideas.RunResponse) bool) ideas.RunResponse {
	a.t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	var got ideas.RunResponse
	for time.Now().Before(deadline) {
		rec := a.call(gloutonRead, http.MethodGet, "/api/runs/"+runID, "")
		if rec.Code != http.StatusOK {
			a.t.Fatalf("GET /runs = %d %s", rec.Code, rec.Body)
		}
		got = ideas.RunResponse{}
		_ = json.Unmarshal(rec.Body.Bytes(), &got)
		if done(got) {
			return got
		}
		time.Sleep(10 * time.Millisecond)
	}
	a.t.Fatalf("the run never reached the state: %+v", got)
	return got
}

func TestAppTokensNeedTheBulleScopeAndAnActor(t *testing.T) {
	a := newAppAPI(t, scripted{}, plan{paid: true})
	space, _ := a.store.EnsureDefaultSpace(context.Background(), 1)
	idea := a.gloutonIdea(space.ID)
	cases := []struct {
		name, token, method, path string
		want                      int
	}{
		{"a read token reads", gloutonRead, http.MethodGet, "/api/ideas?app=glouton", http.StatusOK},
		{"a read token cannot run", gloutonRead, http.MethodPost, "/api/ideas/" + idea.ID + "/run", http.StatusForbidden},
		{"another app's scope cannot read", otherScope, http.MethodGet, "/api/ideas?app=glouton", http.StatusForbidden},
		{"another app's scope cannot run", otherScope, http.MethodPost, "/api/ideas/" + idea.ID + "/run", http.StatusForbidden},
		{"a token acting for nobody is refused", noActor, http.MethodGet, "/api/ideas?app=glouton", http.StatusForbidden},
		{"an unknown token is refused", "x.y.z", http.MethodGet, "/api/ideas", http.StatusUnauthorized},
	}
	for _, c := range cases {
		if rec := a.call(c.token, c.method, c.path, ""); rec.Code != c.want {
			t.Errorf("%s: %d %s, want %d", c.name, rec.Code, rec.Body, c.want)
		}
	}
	if got, _ := a.store.GetIdea(context.Background(), 1, idea.ID); got.Status != store.IdeaNew {
		t.Fatalf("a refused token changed the idea: %s", got.Status)
	}
}

func TestListIsScopedToTheAppAndTheCallersSpace(t *testing.T) {
	a := newAppAPI(t, scripted{}, plan{paid: true})
	ctx := context.Background()
	space, _ := a.store.EnsureDefaultSpace(ctx, 1)
	shared, _ := a.store.CreateSpace(ctx, store.Space{UserID: 1, Name: "Facile", RegistreSpaceID: "reg-9"})
	lead := a.gloutonIdea(shared.ID)
	mail, _ := a.store.CreateIdea(ctx, newIdeaIn(shared.ID, "courrier:5/email:1/reply"))
	a.gloutonIdea(space.ID)
	a.store.EnsureDefaultSpace(ctx, 2)
	stranger, _ := a.store.CreateSpace(ctx, store.Space{UserID: 2, Name: "Theirs", RegistreSpaceID: "reg-7"})

	var page ideas.ListResponse
	rec := a.call(gloutonRead, http.MethodGet, "/api/ideas?app=glouton&space_id=reg-9", "")
	_ = json.Unmarshal(rec.Body.Bytes(), &page)
	if rec.Code != http.StatusOK || len(page.Ideas) != 1 || page.Ideas[0].ID != lead.ID || page.Ideas[0].App != "glouton" {
		t.Fatalf("glouton ideas of the Registre space = %d %s", rec.Code, rec.Body)
	}
	page = ideas.ListResponse{}
	_ = json.Unmarshal(a.call(gloutonRead, http.MethodGet, "/api/ideas?space_id="+shared.ID, "").Body.Bytes(), &page)
	if len(page.Ideas) != 2 {
		t.Fatalf("Bulle's own list lost an app's idea: %+v", page.Ideas)
	}
	page = ideas.ListResponse{}
	_ = json.Unmarshal(a.call(gloutonRead, http.MethodGet, "/api/ideas?app=courrier&space_id=reg-9", "").Body.Bytes(), &page)
	if len(page.Ideas) != 1 || page.Ideas[0].ID != mail.ID {
		t.Fatalf("courrier ideas = %+v", page.Ideas)
	}
	if rec := a.call(gloutonRead, http.MethodGet, "/api/ideas?app=nope", ""); rec.Code != http.StatusBadRequest {
		t.Fatalf("an unknown app = %d", rec.Code)
	}
	for _, other := range []string{"reg-7", stranger.ID, "missing"} {
		if rec := a.call(gloutonRead, http.MethodGet, "/api/ideas?app=glouton&space_id="+other, ""); rec.Code != http.StatusNotFound {
			t.Fatalf("space %s that is not the caller's = %d", other, rec.Code)
		}
	}
	if rec := a.call(gloutonWrite, http.MethodPost, "/api/ideas/"+lead.ID+"/run?space_id="+space.ID, ""); rec.Code != http.StatusNotFound {
		t.Fatalf("running an idea of another space than the active one = %d", rec.Code)
	}
}

func newIdeaIn(spaceID, fingerprint string) store.Idea {
	return store.Idea{
		UserID: 1, SpaceID: spaceID, BatchID: "b", Kind: "mail", Title: "Reply", Why: "Asked", Prompt: "Reply",
		Fingerprint: fingerprint, ExpiresAt: time.Now().Add(time.Hour),
	}
}

func TestABackgroundRunWaitsForAnApprovalAnsweredThroughTheAPI(t *testing.T) {
	backend := scripted{turns: []nacelletest.Turn{
		{Calls: []nacelle.ToolCall{{ID: "c1", Name: "opus_create_task", Input: json.RawMessage(`{"title":"Call Acme"}`)}}},
		{Text: "Created the task Call Acme."},
	}}
	a := newAppAPI(t, backend, plan{paid: true})
	ctx := context.Background()
	space, _ := a.store.EnsureDefaultSpace(ctx, 1)
	idea := a.gloutonIdea(space.ID)

	rec := a.call(gloutonWrite, http.MethodPost, "/api/ideas/"+idea.ID+"/run?space_id="+space.ID, "")
	var started ideas.RunStartResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &started)
	if rec.Code != http.StatusAccepted || started.RunID == "" || started.SessionID == "" || started.IdeaID != idea.ID {
		t.Fatalf("run = %d %s", rec.Code, rec.Body)
	}
	waiting := a.pollRun(started.RunID, func(r ideas.RunResponse) bool { return r.Status == runner.RunWaiting })
	if len(waiting.PendingApprovals) != 1 || waiting.PendingApprovals[0].CallID != "c1" || waiting.PendingApprovals[0].Tool != "opus_create_task" {
		t.Fatalf("pending approvals = %+v", waiting.PendingApprovals)
	}
	pending := waiting.PendingApprovals[0]
	if !strings.Contains(string(pending.Args), "Call Acme") || pending.DecidePath != "/api/runs/"+started.RunID+"/approvals/c1" {
		t.Fatalf("approval card = %+v", pending)
	}
	if len(waiting.Outputs) != 0 {
		t.Fatalf("an unapproved call left an output: %+v", waiting.Outputs)
	}
	if rec := a.call(gloutonRead, http.MethodPost, pending.DecidePath, `{"approve":true}`); rec.Code != http.StatusForbidden {
		t.Fatalf("a read token approved = %d", rec.Code)
	}
	if rec := a.call(gloutonWrite, http.MethodPost, pending.DecidePath, `{"approve":true}`); rec.Code != http.StatusOK {
		t.Fatalf("approve = %d %s", rec.Code, rec.Body)
	}
	done := a.pollRun(started.RunID, func(r ideas.RunResponse) bool { return r.Status == store.RunFinished })
	if done.Summary != "Created the task Call Acme." || len(done.PendingApprovals) != 0 || done.FinishedAt == nil {
		t.Fatalf("finished run = %+v", done)
	}
	if len(done.Outputs) != 1 || done.Outputs[0].ID != "t-42" || done.Outputs[0].URL != "https://opus.facile.studio/tasks/t-42" || done.Outputs[0].App != "opus" {
		t.Fatalf("outputs = %+v", done.Outputs)
	}
	if done.ContinuePath != "/s/"+started.SessionID {
		t.Fatalf("continue path = %s", done.ContinuePath)
	}
	chat, err := a.store.GetSession(ctx, 1, started.SessionID)
	if err != nil || chat.Channel != store.ChannelWeb || chat.SpaceID != space.ID {
		t.Fatalf("the run left no chat to continue in Bulle: %+v %v", chat, err)
	}
	if got, _ := a.store.GetIdea(ctx, 1, idea.ID); got.Status != store.IdeaDone || got.SessionID != started.SessionID {
		t.Fatalf("idea after run = %+v", got)
	}
	again := a.call(gloutonWrite, http.MethodPost, "/api/ideas/"+idea.ID+"/run", "")
	if again.Code != http.StatusOK || !strings.Contains(again.Body.String(), started.RunID) {
		t.Fatalf("running a done idea again = %d %s", again.Code, again.Body)
	}
	if rec := a.call(gloutonRead, http.MethodGet, "/api/runs/unknown", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("an unknown run = %d", rec.Code)
	}
}

func TestAFreePlanReadsButCannotGenerateOrRun(t *testing.T) {
	a := newAppAPI(t, scripted{}, plan{paid: false})
	ctx := context.Background()
	space, _ := a.store.EnsureDefaultSpace(ctx, 1)
	idea := a.gloutonIdea(space.ID)
	if rec := a.call(gloutonRead, http.MethodGet, "/api/ideas?app=glouton", ""); rec.Code != http.StatusOK {
		t.Fatalf("a free plan cannot read = %d", rec.Code)
	}
	for _, path := range []string{"/api/ideas/generate?app=glouton", "/api/ideas/" + idea.ID + "/run"} {
		rec := a.call(gloutonWrite, http.MethodPost, path, `{}`)
		var body struct {
			Error struct {
				Code       string `json:"code"`
				UpgradeURL string `json:"upgrade_url"`
			} `json:"error"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		if rec.Code != http.StatusPaymentRequired || body.Error.Code != "payment_required" || body.Error.UpgradeURL != "https://registre.facile.studio/spaces" {
			t.Fatalf("%s on a free plan = %d %s", path, rec.Code, rec.Body)
		}
	}
	if chats, _ := a.store.ListSessions(ctx, 1, ""); len(chats) != 0 {
		t.Fatalf("a refused run left %d chats", len(chats))
	}
	if got, _ := a.store.GetIdea(ctx, 1, idea.ID); got.Status != store.IdeaNew {
		t.Fatalf("a refused run changed the idea: %s", got.Status)
	}
}

func TestAnAppsGenerateIsLimitedPerUserAcrossSpaces(t *testing.T) {
	a := newAppAPI(t, scripted{}, plan{paid: true})
	ctx := context.Background()
	home, _ := a.store.EnsureDefaultSpace(ctx, 1)
	work, _ := a.store.CreateSpace(ctx, store.Space{UserID: 1, Name: "Work"})
	wait := func(spaceID, app string) {
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			if _, running, _ := a.store.RunningIdeaBatch(ctx, 1, spaceID, app); !running {
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
		t.Fatal("the batch never finished")
	}
	for i, spaceID := range []string{home.ID, work.ID, home.ID} {
		rec := a.call(gloutonWrite, http.MethodPost, "/api/ideas/generate?app=glouton&space_id="+spaceID, `{}`)
		if rec.Code != http.StatusAccepted {
			t.Fatalf("glouton generate %d = %d %s", i+1, rec.Code, rec.Body)
		}
		wait(spaceID, "glouton")
	}
	if rec := a.call(gloutonWrite, http.MethodPost, "/api/ideas/generate?app=glouton&space_id="+work.ID, `{}`); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("glouton generate past the cap = %d %s", rec.Code, rec.Body)
	}
	if rec := a.call(gloutonWrite, http.MethodPost, "/api/ideas/generate?app=opus&space_id="+work.ID, `{}`); rec.Code != http.StatusAccepted {
		t.Fatalf("another app's generate = %d %s", rec.Code, rec.Body)
	}
	wait(work.ID, "opus")
	if rec := a.call(gloutonWrite, http.MethodPost, "/api/ideas/generate", `{"space_id":"`+home.ID+`"}`); rec.Code != http.StatusAccepted {
		t.Fatalf("Bulle's own generate = %d %s", rec.Code, rec.Body)
	}
	if rec := a.call(gloutonWrite, http.MethodPost, "/api/ideas/generate?app=bulle", `{}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("generate for app bulle = %d", rec.Code)
	}
	batches := 0
	for _, app := range []string{"glouton", "opus"} {
		for _, spaceID := range []string{home.ID, work.ID} {
			if last, ok, _ := a.store.LastIdeaBatch(ctx, 1, spaceID, app); ok && last.App == app {
				batches++
			}
		}
	}
	if batches != 3 {
		t.Fatalf("finished app batches = %d", batches)
	}
}
