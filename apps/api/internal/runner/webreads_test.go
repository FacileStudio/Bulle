package runner_test

import (
	"context"
	"sync/atomic"
	"testing"

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

func TestWebReadsRunUnasked(t *testing.T) {
	for _, tool := range []string{"sonar_fetch", "browser_navigate"} {
		f, runID, fetched := fetchRun(t, tool, "what is new on example.com?", "https://news.example/today")
		if run := f.waitDone(t, runID); run.Status != store.RunFinished {
			t.Fatalf("%s run: %+v", tool, run)
		}
		for _, kind := range f.eventTypes(t, runID) {
			if kind == "INTERRUPT" {
				t.Fatalf("%s asked for approval", tool)
			}
		}
		if fetched.Load() != 1 {
			t.Fatalf("%s fetched %d times", tool, fetched.Load())
		}
	}
}
