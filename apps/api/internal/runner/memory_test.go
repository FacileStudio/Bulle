package runner_test

import (
	"context"
	"encoding/json"
	"errors"
	"iter"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/nacelle"
	"github.com/FacileStudio/porte"
)

type fakeMycelium struct {
	srv     *httptest.Server
	profile string
	status  int

	mu     sync.Mutex
	spaces []string
}

func newFakeMycelium(t *testing.T, profile string, status int) *fakeMycelium {
	t.Helper()
	f := &fakeMycelium{profile: profile, status: status}
	f.srv = httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeMycelium) serve(rw http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/.well-known/openid-configuration":
		_ = json.NewEncoder(rw).Encode(map[string]string{"token_endpoint": f.srv.URL + "/token"})
	case "/token":
		_ = json.NewEncoder(rw).Encode(map[string]any{"access_token": "delegated", "expires_in": 300})
	case "/api/sync/files/memory/" + suite.ProfilePage:
		f.mu.Lock()
		f.spaces = append(f.spaces, r.URL.Query().Get("space_id"))
		f.mu.Unlock()
		if f.status != http.StatusOK {
			rw.WriteHeader(f.status)
			return
		}
		_, _ = rw.Write([]byte(f.profile))
	default:
		rw.WriteHeader(http.StatusNotFound)
	}
}

func (f *fakeMycelium) profileReads() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.spaces)
}

func (f *fakeMycelium) ListByUser(_ context.Context, userID int64) ([]porte.StoredIdentity, error) {
	return []porte.StoredIdentity{{UserID: userID, Provider: f.srv.URL, Tokens: porte.TokenSet{AccessToken: "user", Expiry: time.Now().Add(time.Hour)}}}, nil
}

func (f *fakeMycelium) Save(context.Context, porte.StoredIdentity) error { return nil }

func (f *fakeMycelium) tune(c *runner.Config) {
	cfg := suite.Config{Issuer: f.srv.URL, ClientID: "bulle", ClientSecret: "s3cret", Apps: map[string]suite.AppConfig{}}
	for _, app := range []string{suite.AppCourrier, suite.AppOpus, suite.AppMycelium} {
		cfg.Apps[app] = suite.AppConfig{URL: f.srv.URL, Audience: "aud-" + app}
	}
	c.Suite = suite.NewGateway(suite.NewTokens(cfg, f, nil), cfg, nil)
	c.Outward = suite.WriteTools()
}

func memorySpace(t *testing.T, f *fixture, myceliumSpace string) store.Space {
	t.Helper()
	ctx := context.Background()
	space, err := f.store.EnsureDefaultSpace(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	space.MyceliumSpaceID = myceliumSpace
	space, err = f.store.UpdateSpace(ctx, space)
	if err != nil {
		t.Fatal(err)
	}
	return space
}

func (p *promptBackend) prompt() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.system
}

func TestProfileIsInjectedAndCapped(t *testing.T) {
	profile := "Ann is a designer who likes short answers.\n" + strings.Repeat("She is working on the new website.\n", 200) + "PROFILE-TAIL"
	mycelium := newFakeMycelium(t, profile, http.StatusOK)
	backend := &promptBackend{}
	f := newFixture(t, backend, nil, mycelium.tune)
	memorySpace(t, f, "myc-home")
	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "hello")
	if err != nil {
		t.Fatal(err)
	}
	f.waitDone(t, runID)
	system := backend.prompt()
	if !strings.Contains(system, "What you know about the user") || !strings.Contains(system, "Ann is a designer who likes short answers.") {
		t.Fatalf("the profile is not in the system prompt: %q", system)
	}
	if strings.Contains(system, "PROFILE-TAIL") {
		t.Fatal("the profile was not capped")
	}
	section := system[strings.Index(system, "Ann is a designer"):]
	if end := strings.Index(section, "is longer; read it with memory_read for the rest]"); end < 0 || end > 4<<10 {
		t.Fatalf("want the profile cut under 4 KB with a note, got the note at %d", end)
	}
	if reads := mycelium.profileReads(); len(reads) != 1 || reads[0] != "myc-home" {
		t.Fatalf("want one profile read in the space's Mycelium space, got %v", reads)
	}
}

func TestMyceliumErrorDoesNotFailARun(t *testing.T) {
	for _, status := range []int{http.StatusNotFound, http.StatusInternalServerError} {
		mycelium := newFakeMycelium(t, "", status)
		backend := &promptBackend{}
		f := newFixture(t, backend, nil, mycelium.tune)
		memorySpace(t, f, "myc-home")
		runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "hello")
		if err != nil {
			t.Fatal(err)
		}
		if run := f.waitDone(t, runID); run.Status != store.RunFinished {
			t.Fatalf("Mycelium answering %d failed the run: %s", status, run.Status)
		}
		if strings.Contains(backend.prompt(), "What you know about the user") {
			t.Fatalf("Mycelium answering %d still injected a profile", status)
		}
	}
}

func TestCronRunsGetNoProfile(t *testing.T) {
	mycelium := newFakeMycelium(t, "Ann", http.StatusOK)
	backend := &promptBackend{}
	f := newFixture(t, backend, nil, mycelium.tune)
	space := memorySpace(t, f, "myc-home")
	session, err := f.store.CreateSession(context.Background(), store.Session{ID: store.NewID(), UserID: 1, SpaceID: space.ID, Channel: store.ChannelCron})
	if err != nil {
		t.Fatal(err)
	}
	runID, err := f.runner.Start(context.Background(), 1, session.ID, "daily digest")
	if err != nil {
		t.Fatal(err)
	}
	f.waitDone(t, runID)
	if reads := mycelium.profileReads(); len(reads) != 0 {
		t.Fatalf("a cron run read the profile: %v", reads)
	}
}

func TestMemoryPassHasOnlyTheMemoryTools(t *testing.T) {
	mycelium := newFakeMycelium(t, "", http.StatusNotFound)
	block := make(chan struct{})
	backend := &ideasBackend{fakeBackend: fakeBackend{block: block}}
	shell := newTool(t, "send_mail", func(string) (string, error) { return "sent", nil })
	f := newFixture(t, backend, []nacelle.Tool{shell}, mycelium.tune)
	space := memorySpace(t, f, "myc-home")
	first := chatWith(t, f, space.ID, "I am Ann.", "Nice to meet you, Ann.")
	second := chatWith(t, f, space.ID, "I moved to Lyon.")
	ctx := context.Background()

	if err := f.runner.MemoryPass(ctx, 1, first.ID); err != nil {
		t.Fatal(err)
	}
	if err := f.runner.MemoryPass(ctx, 1, second.ID); !errors.Is(err, runner.ErrMemoryBusy) {
		t.Fatalf("want a second pass for the same user refused while one runs, got %v", err)
	}
	waitFor(t, "the pass to call the model", func() bool { return len(backend.seen()) > 0 })
	names := []string{}
	for _, tool := range backend.seen() {
		names = append(names, tool.Spec().Name)
	}
	slices.Sort(names)
	want := slices.Sorted(slices.Values(suite.MemoryTools()))
	if !slices.Equal(names, want) {
		t.Fatalf("want exactly the memory tools, got %v", names)
	}
	close(block)
	waitFor(t, "the chat to be marked seen", func() bool {
		seen, _ := f.store.GetSession(ctx, 1, first.ID)
		return seen.MemorySeenSeq == 2
	})
	waitFor(t, "the pass to let the user go", func() bool {
		return f.runner.MemoryPass(ctx, 1, second.ID) == nil
	})
	listed, _ := f.store.ListSessions(ctx, 1, "")
	for _, session := range listed {
		if session.Channel == store.ChannelMemory {
			t.Fatal("a memory pass session shows in the session list")
		}
	}
	if f.computers.ensured.Load() != 0 {
		t.Fatal("a memory pass booted the computer")
	}
}

func TestMemoryPassOnlyLearnsFromChats(t *testing.T) {
	mycelium := newFakeMycelium(t, "", http.StatusNotFound)
	f := newFixture(t, &fakeBackend{}, nil, mycelium.tune)
	space := memorySpace(t, f, "myc-home")
	session, err := f.store.CreateSession(context.Background(), store.Session{ID: store.NewID(), UserID: 1, SpaceID: space.ID, Channel: store.ChannelCron})
	if err != nil {
		t.Fatal(err)
	}
	if err := f.runner.MemoryPass(context.Background(), 1, session.ID); !errors.Is(err, runner.ErrNotAChat) {
		t.Fatalf("want a cron session refused, got %v", err)
	}
	plain := newFixture(t, &fakeBackend{}, nil, nil)
	if err := plain.runner.MemoryPass(context.Background(), 1, plain.session.ID); !errors.Is(err, runner.ErrMemoryOff) {
		t.Fatalf("want no pass without Mycelium, got %v", err)
	}
}

type failingBackend struct{ fakeBackend }

func (failingBackend) Stream(context.Context, nacelle.Request) iter.Seq2[nacelle.Event, error] {
	return func(yield func(nacelle.Event, error) bool) { yield(nacelle.Event{}, errors.New("model refused")) }
}

func TestFailedMemoryPassLeavesTheChatToLearnAgain(t *testing.T) {
	mycelium := newFakeMycelium(t, "", http.StatusNotFound)
	f := newFixture(t, &failingBackend{}, nil, mycelium.tune)
	space := memorySpace(t, f, "myc-home")
	chat := chatWith(t, f, space.ID, "I am Ann.", "Nice to meet you, Ann.")
	ctx := context.Background()
	if err := f.runner.MemoryPass(ctx, 1, chat.ID); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "the pass to let the user go", func() bool {
		return f.runner.MemoryPass(ctx, 1, chat.ID) == nil
	})
	if seen, _ := f.store.GetSession(ctx, 1, chat.ID); seen.MemorySeenSeq != 0 {
		t.Fatalf("a failed pass marked the chat seen up to %d", seen.MemorySeenSeq)
	}
}
