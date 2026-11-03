package runner_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/nacelle"
	"github.com/FacileStudio/porte"
)

type fakeSablier struct {
	srv  *httptest.Server
	fail bool

	mu    sync.Mutex
	posts map[string][]map[string]any
}

func newFakeSablier(t *testing.T, fail bool) *fakeSablier {
	t.Helper()
	f := &fakeSablier{fail: fail, posts: map[string][]map[string]any{}}
	f.srv = httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeSablier) serve(rw http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/.well-known/openid-configuration":
		_ = json.NewEncoder(rw).Encode(map[string]string{"token_endpoint": f.srv.URL + "/token"})
	case "/token":
		_ = json.NewEncoder(rw).Encode(map[string]any{"access_token": "delegated", "expires_in": 300})
	default:
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		f.mu.Lock()
		f.posts[r.URL.Path] = append(f.posts[r.URL.Path], body)
		f.mu.Unlock()
		if f.fail {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = rw.Write([]byte(`{}`))
	}
}

func (f *fakeSablier) sent(path string) []map[string]any {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]map[string]any(nil), f.posts[path]...)
}

func (f *fakeSablier) ListByUser(_ context.Context, userID int64) ([]porte.StoredIdentity, error) {
	return []porte.StoredIdentity{{UserID: userID, Provider: f.srv.URL, Tokens: porte.TokenSet{AccessToken: "user", Expiry: time.Now().Add(time.Hour)}}}, nil
}

func (f *fakeSablier) Save(context.Context, porte.StoredIdentity) error { return nil }

func (f *fakeSablier) tune(c *runner.Config) {
	cfg := suite.Config{Issuer: f.srv.URL, ClientID: "bulle", ClientSecret: "s3cret", Apps: map[string]suite.AppConfig{
		suite.AppSablier: {URL: f.srv.URL, Audience: "aud-sablier"},
	}}
	c.Suite = suite.NewGateway(suite.NewTokens(cfg, f, nil), cfg, nil)
	c.TimeTracking = true
	var mu sync.Mutex
	clock := time.Date(2026, 11, 3, 9, 0, 0, 0, time.UTC)
	c.Now = func() time.Time {
		mu.Lock()
		defer mu.Unlock()
		clock = clock.Add(time.Minute)
		return clock
	}
}

func TestRunIsBookedToTheAgentWithoutTheApprovalWait(t *testing.T) {
	sablier := newFakeSablier(t, false)
	send := newTool(t, "send_mail", func(string) (string, error) { return "sent", nil })
	backend := &fakeBackend{calls: []call{{id: "m1", name: "send_mail", input: `{"text":"to bob"}`}}}
	f := newFixture(t, backend, []nacelle.Tool{send}, func(c *runner.Config) {
		sablier.tune(c)
		c.Outward = []string{"send_mail"}
	})
	if err := f.store.SaveAssistantProfile(context.Background(), store.AssistantProfile{UserID: 1, Name: "Jarvis", AvatarURL: "https://blobatar.dev/avatar/Jarvis?gen=2"}); err != nil {
		t.Fatal(err)
	}

	runID := f.runApproved(t, "mail bob", "m1")
	waitFor(t, "the agent session", func() bool { return len(sablier.sent("/api/agents/sessions")) == 1 })

	got := sablier.sent("/api/agents/sessions")[0]
	if got["agent_key"] != "bulle" || got["name"] != "Jarvis" || got["facile_id"] != "bulle_run_"+runID || got["cowork"] != true || got["description"] != "mail bob" {
		t.Fatalf("unexpected agent session: %v", got)
	}
	if got["started_at"] != "2026-11-03T09:01:00Z" || got["stopped_at"] != "2026-11-03T09:06:00Z" || got["active_seconds"] != float64(240) {
		t.Fatalf("want five minutes of wall time minus one minute of approval wait, got %v", got)
	}
	sablier.assertOneHeartbeat(t, f.session.ID)
}

func (f *fixture) runApproved(t *testing.T, text, callID string) string {
	t.Helper()
	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, text)
	if err != nil {
		t.Fatal(err)
	}
	f.waitEvent(t, runID, "INTERRUPT")
	if err := f.runner.Decide(context.Background(), 1, runID, callID, true); err != nil {
		t.Fatal(err)
	}
	f.waitDone(t, runID)
	return runID
}

func (f *fakeSablier) assertOneHeartbeat(t *testing.T, sessionID string) {
	t.Helper()
	beats := f.sent("/api/heartbeats")
	if len(beats) != 1 {
		t.Fatalf("want one heartbeat, got %v", beats)
	}
	beat := beats[0]["heartbeats"].([]any)[0].(map[string]any)
	if beat["source"] != "bulle" || beat["entity"] != "bulle:session/"+sessionID || beat["project"] != "" {
		t.Fatalf("unexpected heartbeat: %v", beat)
	}
}

func TestFailingSablierLeavesTheRunFinished(t *testing.T) {
	sablier := newFakeSablier(t, true)
	f := newFixture(t, &fakeBackend{}, nil, sablier.tune)
	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "hello")
	if err != nil {
		t.Fatal(err)
	}
	if run := f.waitDone(t, runID); run.Status != store.RunFinished {
		t.Fatalf("want a finished run, got %s", run.Status)
	}
	waitFor(t, "the agent session attempt", func() bool { return len(sablier.sent("/api/agents/sessions")) == 1 })
}

func TestRunThatFailedBeforeTheModelIsNotBooked(t *testing.T) {
	sablier := newFakeSablier(t, false)
	f := newFixture(t, &fakeBackend{}, nil, func(c *runner.Config) {
		sablier.tune(c)
		c.Backends = func(string, string) (nacelle.Backend, error) { return nil, errors.New("no key") }
	})
	cron, err := f.store.CreateSession(context.Background(), store.Session{ID: store.NewID(), UserID: 1, Channel: store.ChannelCron})
	if err != nil {
		t.Fatal(err)
	}
	runID, err := f.runner.Start(context.Background(), 1, cron.ID, "hello")
	if err != nil {
		t.Fatal(err)
	}
	if run := f.waitDone(t, runID); run.Status != store.RunError {
		t.Fatalf("want an errored run, got %s", run.Status)
	}
	time.Sleep(100 * time.Millisecond)
	if posts := len(sablier.sent("/api/agents/sessions")) + len(sablier.sent("/api/heartbeats")); posts != 0 {
		t.Fatalf("want nothing sent to Sablier, got %d calls", posts)
	}
}
