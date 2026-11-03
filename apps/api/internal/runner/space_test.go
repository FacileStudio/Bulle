package runner_test

import (
	"context"
	"iter"
	"strings"
	"sync"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

type promptBackend struct {
	mu     sync.Mutex
	system string
}

func (p *promptBackend) Name() string { return "prompt" }

func (p *promptBackend) Capabilities() nacelle.Capabilities {
	return nacelle.Capabilities{Thinking: true}
}

func (p *promptBackend) CountTokens(context.Context, nacelle.Request) (int64, error) { return 0, nil }

func (p *promptBackend) Stream(_ context.Context, req nacelle.Request) iter.Seq2[nacelle.Event, error] {
	p.mu.Lock()
	p.system = req.System
	p.mu.Unlock()
	return func(yield func(nacelle.Event, error) bool) {
		if yield(nacelle.Event{Kind: nacelle.KindText, Text: "ok"}, nil) {
			yield(nacelle.Event{Kind: nacelle.KindDone, Stop: nacelle.StopEnd}, nil)
		}
	}
}

func TestSystemPromptNamesTheSessionSpace(t *testing.T) {
	backend := &promptBackend{}
	f := newFixture(t, backend, nil, nil)
	ctx := context.Background()
	work, err := f.store.CreateSpace(ctx, store.Space{UserID: 1, Name: "Work"})
	if err != nil {
		t.Fatal(err)
	}
	session, err := f.store.CreateSession(ctx, store.Session{ID: store.NewID(), UserID: 1, SpaceID: work.ID, Channel: store.ChannelWeb})
	if err != nil {
		t.Fatal(err)
	}
	runID, err := f.runner.Start(ctx, 1, session.ID, "hello")
	if err != nil {
		t.Fatal(err)
	}
	f.waitDone(t, runID)
	backend.mu.Lock()
	defer backend.mu.Unlock()
	if !strings.Contains(backend.system, "You are working in the Work space") {
		t.Fatalf("the system prompt does not name the space: %q", backend.system)
	}
}
