package runner_test

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/nacelle"
)

type ideasBackend struct {
	fakeBackend
	mu    sync.Mutex
	tools []nacelle.Tool
}

func (b *ideasBackend) Stream(ctx context.Context, req nacelle.Request) iter.Seq2[nacelle.Event, error] {
	b.mu.Lock()
	b.tools = req.Tools
	b.mu.Unlock()
	return b.fakeBackend.Stream(ctx, req)
}

func (b *ideasBackend) seen() []nacelle.Tool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return slices.Clone(b.tools)
}

func proposals(fingerprints ...string) []call {
	out := make([]call, len(fingerprints))
	for i, fp := range fingerprints {
		input := fmt.Sprintf(`{"kind":"task","title":"Idea %d","why":"It is due","prompt":"Do it","fingerprint":%q}`, i, fp)
		out[i] = call{id: fmt.Sprintf("p%d", i), name: runner.ProposeToolName, input: input}
	}
	return out
}

func fullSuite(c *runner.Config) {
	cfg := suite.Config{Issuer: "http://127.0.0.1:1", ClientID: "bulle", ClientSecret: "s", Apps: map[string]suite.AppConfig{}}
	for _, app := range []string{suite.AppCourrier, suite.AppAgenda, suite.AppOpus, suite.AppSablier, suite.AppMycelium, suite.AppNuage, suite.AppGlouton} {
		cfg.Apps[app] = suite.AppConfig{URL: "http://127.0.0.1:1", Audience: "aud-" + app}
	}
	c.Suite = suite.NewGateway(suite.NewTokens(cfg, nil, nil), cfg, nil)
	c.Outward = append(suite.WriteTools(), "browser_confirm", "send_mail")
}

func (f *fixture) waitBatch(t *testing.T, spaceID string) {
	t.Helper()
	waitFor(t, "the ideas batch to end", func() bool {
		_, running, _ := f.store.RunningIdeaBatch(context.Background(), 1, spaceID, "")
		return !running
	})
}

func (f *fixture) space(t *testing.T) store.Space {
	t.Helper()
	space, err := f.store.EnsureDefaultSpace(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	return space
}

func TestIdeasAgentOnlyReadsAndProposes(t *testing.T) {
	backend := &ideasBackend{fakeBackend: fakeBackend{calls: proposals("opus:task:1/due")}}
	shell := newTool(t, "send_mail", func(string) (string, error) { return "sent", nil })
	f := newFixture(t, backend, []nacelle.Tool{shell}, fullSuite)
	space := f.space(t)
	if _, err := f.runner.GenerateIdeas(context.Background(), 1, space.ID, store.BatchManual); err != nil {
		t.Fatal(err)
	}
	f.waitBatch(t, space.ID)
	gated := map[string]bool{"send_mail": true, "browser_confirm": true}
	for _, name := range suite.WriteTools() {
		gated[name] = true
	}
	names := []string{}
	for _, tool := range backend.seen() {
		names = append(names, tool.Spec().Name)
		if tool.Spec().Name != runner.ProposeToolName && !tool.Spec().ReadOnly {
			t.Errorf("%s is not read-only", tool.Spec().Name)
		}
		if gated[tool.Spec().Name] {
			t.Errorf("%s reaches the approval gate", tool.Spec().Name)
		}
	}
	if !slices.Contains(names, runner.ProposeToolName) || !slices.Contains(names, "courrier_list_inbox") {
		t.Fatalf("the ideas agent lacks its tools: %v", names)
	}
	if f.computers.ensured.Load() != 0 {
		t.Fatal("an ideas batch booted the computer")
	}
	if active, _ := f.store.ListActiveIdeas(context.Background(), 1, space.ID, "", time.Now()); len(active) != 1 {
		t.Fatalf("want the proposed idea stored, got %d", len(active))
	}
	if listed, _ := f.store.ListSessions(context.Background(), 1, ""); len(listed) != 1 {
		t.Fatalf("the ideas session shows in the session list: %d sessions", len(listed))
	}
}

func TestGenerateIdeasIsBusyWhileABatchRuns(t *testing.T) {
	block := make(chan struct{})
	backend := &fakeBackend{block: block, calls: proposals("opus:task:2/due")}
	f := newFixture(t, backend, nil, nil)
	ctx := context.Background()
	space := f.space(t)
	stale, _ := f.store.CreateIdea(ctx, store.Idea{UserID: 1, SpaceID: space.ID, BatchID: "old", Kind: "task", Title: "Old",
		Why: "w", Prompt: "p", Fingerprint: "old/new", ExpiresAt: time.Now().Add(time.Hour)})
	kept, _ := f.store.CreateIdea(ctx, store.Idea{UserID: 1, SpaceID: space.ID, BatchID: "old", Kind: "task", Title: "Kept",
		Why: "w", Prompt: "p", Fingerprint: "old/saved", Status: store.IdeaSaved, ExpiresAt: time.Now().Add(time.Hour)})
	if _, err := f.runner.GenerateIdeas(ctx, 1, space.ID, store.BatchManual); err != nil {
		t.Fatal(err)
	}
	if _, err := f.runner.GenerateIdeas(ctx, 1, space.ID, store.BatchManual); !errors.Is(err, runner.ErrIdeasBusy) {
		t.Fatalf("a second batch while one runs: %v", err)
	}
	close(block)
	f.waitBatch(t, space.ID)
	if got, _ := f.store.GetIdea(ctx, 1, stale.ID); got.Status != store.IdeaExpired {
		t.Errorf("the older new idea is %s, want expired", got.Status)
	}
	if got, _ := f.store.GetIdea(ctx, 1, kept.ID); got.Status != store.IdeaSaved {
		t.Errorf("the saved idea is %s, want saved", got.Status)
	}
	if last, ok, _ := f.store.LastIdeaBatch(ctx, 1, space.ID, ""); !ok || last.RunID == "" {
		t.Fatalf("the batch did not finish with its run: %+v", last)
	}
}

func TestIdeasPushOnlyForADailyBatchWithIdeas(t *testing.T) {
	cases := []struct {
		name    string
		trigger string
		calls   []call
		push    bool
		want    []string
	}{
		{"daily with ideas", store.BatchDaily, proposals("a", "b"), true, []string{"2 ideas for Personal|ideas"}},
		{"daily push off", store.BatchDaily, proposals("a"), false, []string{}},
		{"manual", store.BatchManual, proposals("a"), true, []string{}},
		{"daily empty", store.BatchDaily, nil, true, []string{}},
	}
	for _, c := range cases {
		sent := &recordedPush{}
		f := newFixture(t, &fakeBackend{calls: c.calls}, nil, func(cfg *runner.Config) { cfg.Push = sent })
		space := f.space(t)
		settings := store.DefaultIdeaSettings(1, space.ID)
		settings.Push = c.push
		if _, err := f.store.SaveIdeaSettings(context.Background(), settings); err != nil {
			t.Fatal(err)
		}
		if _, err := f.runner.GenerateIdeas(context.Background(), 1, space.ID, c.trigger); err != nil {
			t.Fatal(err)
		}
		f.waitBatch(t, space.ID)
		if got := sent.bodies(); !slices.Equal(got, c.want) {
			t.Errorf("%s: pushes = %v, want %v", c.name, got, c.want)
		}
	}
}
