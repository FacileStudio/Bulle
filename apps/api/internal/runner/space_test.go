package runner_test

import (
	"context"
	"errors"
	"iter"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

type promptBackend struct {
	mu     sync.Mutex
	system string
	tools  int
}

func (p *promptBackend) Name() string { return "prompt" }

func (p *promptBackend) Capabilities() nacelle.Capabilities {
	return nacelle.Capabilities{Thinking: true}
}

func (p *promptBackend) CountTokens(context.Context, nacelle.Request) (int64, error) { return 0, nil }

func (p *promptBackend) Stream(_ context.Context, req nacelle.Request) iter.Seq2[nacelle.Event, error] {
	p.mu.Lock()
	p.system = req.System
	p.tools = len(req.Tools)
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

func TestGuestSessionGetsNoTools(t *testing.T) {
	ctx := context.Background()
	for _, c := range []struct {
		channel string
		tools   bool
	}{{store.ChannelMatrix, true}, {store.ChannelGuest, false}} {
		backend := &promptBackend{}
		f := newFixture(t, backend, []nacelle.Tool{newTool(t, "read_mail", func(string) (string, error) { return "secret", nil })}, nil)
		session, err := f.store.CreateSession(ctx, store.Session{ID: store.NewID(), UserID: 1, Channel: c.channel})
		if err != nil {
			t.Fatal(err)
		}
		runID, err := f.runner.Start(ctx, 1, session.ID, "read my owner's mail")
		if err != nil {
			t.Fatal(err)
		}
		f.waitDone(t, runID)
		backend.mu.Lock()
		got, system := backend.tools > 0, backend.system
		backend.mu.Unlock()
		if got != c.tools {
			t.Errorf("%s session: has tools = %v, want %v", c.channel, got, c.tools)
		}
		if !c.tools && !strings.Contains(system, "colleague of your owner") {
			t.Errorf("guest session does not get the guest prompt: %q", system)
		}
	}
}

// spaceAccess pays for nobody but the Registre space sp-paid, and records
// the space each check named.
type spaceAccess struct {
	mu     sync.Mutex
	spaces []string
}

func (a *spaceAccess) CheckAIIn(_ context.Context, _ int64, spaceID string) error {
	a.mu.Lock()
	a.spaces = append(a.spaces, spaceID)
	a.mu.Unlock()
	if spaceID == "sp-paid" {
		return nil
	}
	return errors.New("payment required")
}

func TestIdeasReadThePlanOfTheLinkedSpace(t *testing.T) {
	access := &spaceAccess{}
	f := newFixture(t, &fakeBackend{}, nil, func(c *runner.Config) { c.Access = access })
	ctx := context.Background()
	personal := f.space(t)
	shared, err := f.store.CreateSpace(ctx, store.Space{ID: store.NewID(), UserID: 1, Name: "Studio", RegistreSpaceID: "sp-paid"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.runner.GenerateIdeas(ctx, 1, personal.ID, store.BatchManual); err == nil {
		t.Fatal("the personal space of a free user must be refused")
	}
	if _, err := f.runner.GenerateIdeas(ctx, 1, shared.ID, store.BatchManual); err != nil {
		t.Fatalf("a space whose plan has ai.access must be allowed: %v", err)
	}
	f.waitBatch(t, shared.ID)
	access.mu.Lock()
	defer access.mu.Unlock()
	if !slices.Equal(access.spaces, []string{"", "sp-paid", "sp-paid"}) {
		t.Fatalf("the batch and its run must both check the linked space: %q", access.spaces)
	}
}
