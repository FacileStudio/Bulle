package runner_test

import (
	"context"
	"iter"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

type appBatchBackend struct {
	ideasBackend
	mu     sync.Mutex
	system string
	first  string
}

func (b *appBatchBackend) Stream(ctx context.Context, req nacelle.Request) iter.Seq2[nacelle.Event, error] {
	b.mu.Lock()
	if b.system == "" {
		b.system = req.System
		for _, message := range req.Messages {
			for _, part := range message.Parts {
				if text, ok := part.(nacelle.Text); ok {
					b.first += text.Text
				}
			}
		}
	}
	b.mu.Unlock()
	return b.ideasBackend.Stream(ctx, req)
}

func TestAnAppBatchReadsOnlyThatAppAndFilesItsIdeasUnderIt(t *testing.T) {
	backend := &appBatchBackend{ideasBackend: ideasBackend{fakeBackend: fakeBackend{calls: proposals("courrier:5/email:9/reply")}}}
	f := newFixture(t, backend, nil, fullSuite)
	space := f.space(t)
	if _, err := f.runner.GenerateAppIdeas(context.Background(), 1, space.ID, "glouton", store.BatchManual); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "the glouton batch to end", func() bool {
		_, running, _ := f.store.RunningIdeaBatch(context.Background(), 1, space.ID, "glouton")
		return !running
	})
	for _, tool := range backend.seen() {
		name := tool.Spec().Name
		if name != runner.ProposeToolName && !strings.HasPrefix(name, "glouton_") {
			t.Errorf("a glouton batch was handed %s", name)
		}
	}
	backend.mu.Lock()
	system, first := backend.system, backend.first
	backend.mu.Unlock()
	if !strings.Contains(system, "for Glouton only") || !strings.Contains(first, "Lead follow-ups") || !strings.Contains(first, "Newly registered competitors") {
		t.Fatalf("the batch lacks Glouton's guidance:\nsystem: %s\nprompt: %s", system, first)
	}
	glouton, _ := f.store.ListActiveIdeas(context.Background(), 1, space.ID, "glouton", time.Now())
	if len(glouton) != 1 || glouton[0].App != "glouton" {
		t.Fatalf("the idea is not filed under glouton: %+v", glouton)
	}
	if courrier, _ := f.store.ListActiveIdeas(context.Background(), 1, space.ID, "courrier", time.Now()); len(courrier) != 0 {
		t.Fatalf("a glouton batch filed a courrier idea: %+v", courrier)
	}
}

func TestAnUnscopedBatchFilesEachIdeaUnderItsFingerprintsApp(t *testing.T) {
	backend := &ideasBackend{fakeBackend: fakeBackend{calls: proposals("opus:task:1/due", "weather:paris")}}
	f := newFixture(t, backend, nil, fullSuite)
	space := f.space(t)
	if _, err := f.runner.GenerateIdeas(context.Background(), 1, space.ID, store.BatchManual); err != nil {
		t.Fatal(err)
	}
	f.waitBatch(t, space.ID)
	apps := map[string]bool{}
	all, _ := f.store.ListActiveIdeas(context.Background(), 1, space.ID, "", time.Now())
	for _, idea := range all {
		apps[idea.App] = true
	}
	if len(all) != 2 || !apps["opus"] || !apps[store.IdeaAppBulle] {
		t.Fatalf("apps = %v", apps)
	}
}
