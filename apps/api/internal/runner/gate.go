package runner

import (
	"context"
	"sync"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

// EventApprovalError tells the owner an approval card could not be put up, so
// the call it was for was refused. It does not end the run.
const EventApprovalError = "APPROVAL_ERROR"

type gate struct {
	outward map[string]bool
	store   store.Store
	emitter *emitter
	runID   string
	session string
	pended  func(callID, tool string)
	decided func(callID string, approved bool)
	now     func() time.Time

	mu        sync.Mutex
	blocking  int
	blockedAt time.Time
	blocked   time.Duration
	waiting   map[string]chan bool
	policies  map[string]approvalPolicy
	grants    map[string]bool
	scopes    map[string]string
	announced map[string]chan struct{}
	edits     editState
	trust     trust
}

func newGate(outward map[string]bool, st store.Store, em *emitter, runID string, now func() time.Time) *gate {
	return &gate{
		now:       now,
		outward:   outward,
		store:     st,
		emitter:   em,
		runID:     runID,
		waiting:   map[string]chan bool{},
		grants:    map[string]bool{},
		scopes:    map[string]string{},
		announced: map[string]chan struct{}{},
	}
}

// approve is the run's nacelle Approve hook. A call that waits for the owner
// is keyed by the id the model gave it. A link past the run's budget shows
// the link card, which the owner may grant for the rest of the run, and so
// may a card the taint raises (see taintScope).
func (g *gate) approve(ctx context.Context, call nacelle.Call) nacelle.Decision {
	kind := g.need(call.Name, call.Input, true)
	if kind == askNo {
		return nacelle.Decision{}
	}
	args, scope, ok := g.describe(ctx, call.Name, call.Input)
	if !ok {
		return nacelle.Decision{Deny: "the call could not be shown to the user for approval"}
	}
	switch kind {
	case askLinks:
		args, scope = linkCard(call.Input), linkScope
	case askTaint:
		scope = g.taintScope(call.Name, scope)
	}
	if g.granted(call.Name, scope) {
		return nacelle.Decision{}
	}
	if err := g.store.CreateApproval(context.WithoutCancel(ctx), g.runID, call.ID); err != nil {
		g.unrecorded(call.ID, call.Name, err)
		return nacelle.Decision{Deny: "the approval could not be recorded"}
	}
	g.remember(call.ID, call.Name, call.Input)
	if !g.hold(ctx, call, card{args: args, scope: scope}) {
		return nacelle.Decision{Deny: "the user did not approve it"}
	}
	return nacelle.Decision{}
}

type card struct {
	args  any
	scope string
}

// hold puts the card up and waits for the owner, or for the run to end.
func (g *gate) hold(ctx context.Context, call nacelle.Call, c card) bool {
	id, name := call.ID, call.Name
	background := context.WithoutCancel(ctx)
	announced := make(chan struct{})
	g.mu.Lock()
	g.scopes[id] = c.scope
	g.announced[id] = announced
	g.mu.Unlock()
	decision, release := g.await(id)
	defer release()
	grantable := g.grantable(name, c.scope)
	g.emitter.emit(EventInterrupt, map[string]any{"call_id": id, "tool": name, "args": c.args, "grantable": grantable})
	if g.pended != nil {
		g.pended(id, name)
	}
	close(announced)
	select {
	case approved := <-decision:
		g.settle(background, name, call.Input, c.scope, approved)
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
