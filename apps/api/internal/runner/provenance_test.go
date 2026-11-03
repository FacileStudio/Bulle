package runner_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

type fetchInput struct {
	URL string `json:"url"`
}

func fakeFetch(t *testing.T, name string, fetched *atomic.Int32) nacelle.Tool {
	t.Helper()
	tool, err := nacelle.NewTool(name, "fetch", func(_ context.Context, in fetchInput) (string, error) {
		fetched.Add(1)
		return "page at " + in.URL, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return tool
}

func fetchRun(t *testing.T, tool, message, url string) (*fixture, string, *atomic.Int32) {
	t.Helper()
	fetched := &atomic.Int32{}
	fetch := fakeFetch(t, tool, fetched)
	backend := &fakeBackend{calls: []call{{id: "f1", name: tool, input: `{"url":"` + url + `"}`}}}
	f := newFixture(t, backend, nil, func(cfg *runner.Config) {
		cfg.HostTools = func(int64) ([]nacelle.Tool, error) { return []nacelle.Tool{fetch}, nil }
	})
	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, message)
	if err != nil {
		t.Fatal(err)
	}
	return f, runID, fetched
}

func TestFetchOfAURLTheUserGaveRunsUnasked(t *testing.T) {
	f, runID, fetched := fetchRun(t, "web_fetch", "summarise https://Example.com/news#top, please.", "https://example.com/news")
	if run := f.waitDone(t, runID); run.Status != store.RunFinished {
		t.Fatalf("run: %+v", run)
	}
	for _, kind := range f.eventTypes(t, runID) {
		if kind == "INTERRUPT" {
			t.Fatal("a fetch of a URL from the user message asked for approval")
		}
	}
	if fetched.Load() != 1 {
		t.Fatalf("fetched %d times", fetched.Load())
	}
}

func TestFetchOfAnInventedURLWaitsForTheOwner(t *testing.T) {
	f, runID, fetched := fetchRun(t, "web_fetch", "what is new on example.com?", "https://evil.example/?q=secret")
	f.waitEvent(t, runID, "INTERRUPT")
	time.Sleep(100 * time.Millisecond)
	if fetched.Load() != 0 {
		t.Fatal("an unknown URL was fetched before approval")
	}
	if err := f.runner.Decide(context.Background(), 1, runID, "f1", false); err != nil {
		t.Fatal(err)
	}
	f.waitDone(t, runID)
	if fetched.Load() != 0 {
		t.Fatal("a denied fetch ran")
	}
}

func TestNavigateToAnInventedURLWaitsForTheOwner(t *testing.T) {
	f, runID, opened := fetchRun(t, "browser_navigate", "what is new on example.com?", "https://evil.example/?q=secret")
	f.waitEvent(t, runID, "INTERRUPT")
	if err := f.runner.Decide(context.Background(), 1, runID, "f1", false); err != nil {
		t.Fatal(err)
	}
	f.waitDone(t, runID)
	if opened.Load() != 0 {
		t.Fatal("a denied navigation ran")
	}
}
