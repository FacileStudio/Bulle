package runner

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

// EventApprovalError tells the owner an approval card could not be put up, so
// the call it was for was refused. It does not end the run.
const EventApprovalError = "APPROVAL_ERROR"

type announced struct {
	id    string
	input string
}

type gate struct {
	outward map[string]bool
	store   store.Store
	emitter *emitter
	runID   string
	pended  func(callID, tool string)
	decided func(callID string, approved bool)
	now     func() time.Time
	handoff chan string

	mu        sync.Mutex
	blocking  int
	blockedAt time.Time
	blocked   time.Duration
	asked     map[string][]announced
	signals   map[string]chan struct{}
	waiting   map[string]chan bool
	orphans   int
	policies  map[string]approvalPolicy
	grants    map[string]bool
	scopes    map[string]string
	announced map[string]chan struct{}
	edits     editState
	wrapped   map[string]bool
	trust     trust
}

func newGate(outward map[string]bool, st store.Store, em *emitter, runID string, now func() time.Time) *gate {
	return &gate{
		now:       now,
		outward:   outward,
		store:     st,
		emitter:   em,
		runID:     runID,
		handoff:   make(chan string, 1),
		asked:     map[string][]announced{},
		signals:   map[string]chan struct{}{},
		waiting:   map[string]chan bool{},
		grants:    map[string]bool{},
		scopes:    map[string]string{},
		announced: map[string]chan struct{}{},
		wrapped:   map[string]bool{},
	}
}

// approve is the run's nacelle Approve hook. An approved call of an editable
// tool hands its call id to the wrapper that runs it next, so the owner's edit
// reaches that call and no twin with the same input.
func (g *gate) approve(ctx context.Context, name string, input json.RawMessage) bool {
	ok, id := g.ask(ctx, name, input)
	g.mu.Lock()
	wrapped := g.wrapped[name]
	g.mu.Unlock()
	if ok && wrapped {
		g.handoff <- id
	}
	return ok
}

func (g *gate) ask(ctx context.Context, name string, input json.RawMessage) (bool, string) {
	if !g.asks(name, input) {
		return true, ""
	}
	id := g.callID(ctx, name, string(input))
	args, scope, ok := g.describe(ctx, name, input)
	if !ok || g.granted(name, scope) {
		return ok, id
	}
	background := context.WithoutCancel(ctx)
	if err := g.store.CreateApproval(background, g.runID, id); err != nil {
		g.unrecorded(id, name, err)
		return false, id
	}
	g.remember(id, name, input)
	return g.hold(ctx, id, name, input, card{args: args, scope: scope}), id
}

type card struct {
	args  any
	scope string
}

// hold puts the card up and waits for the owner, or for the run to end.
func (g *gate) hold(ctx context.Context, id, name string, input json.RawMessage, c card) bool {
	background := context.WithoutCancel(ctx)
	announced := make(chan struct{})
	g.mu.Lock()
	g.scopes[id] = c.scope
	g.announced[id] = announced
	g.mu.Unlock()
	decision, release := g.await(id)
	defer release()
	grantable := c.scope != "" && g.takesGrant(name)
	g.emitter.emit(EventInterrupt, map[string]any{"call_id": id, "tool": name, "args": c.args, "grantable": grantable})
	if g.pended != nil {
		g.pended(id, name)
	}
	close(announced)
	select {
	case approved := <-decision:
		g.settle(background, name, input, c.scope, approved)
		return approved
	case <-ctx.Done():
		g.abandon(background, id)
		g.mu.Lock()
		delete(g.scopes, id)
		delete(g.announced, id)
		g.mu.Unlock()
		return false
	}
}

// unrecorded refuses a call whose approval could not be stored, and says so in
// the log and on the run's stream rather than as a silent denial.
func (g *gate) unrecorded(id, name string, err error) {
	g.emitter.logger.Error("recording an approval", "run", g.runID, "call", id, "tool", name, "error", err)
	g.emitter.emit(EventApprovalError, map[string]any{
		"call_id": id, "tool": name, "message": "The approval could not be recorded, so " + name + " did not run.",
	})
}

func (g *gate) pending(callID string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	_, ok := g.waiting[callID]
	return ok
}
