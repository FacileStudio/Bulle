package runner

import (
	"context"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

// decide records the owner's answer and releases the waiting call. The
// approval.decided event goes out before the call resumes, so no client sees
// the run finish ahead of the decision that let it go on.
func (g *gate) decide(ctx context.Context, callID string, approve bool, by string, edit Edit, remember bool) error {
	if !g.pending(callID) {
		return ErrNoPendingApproval
	}
	if approve && !edit.empty() {
		if err := g.revise(callID, edit); err != nil {
			return err
		}
	}
	if err := g.record(ctx, callID, approve, by); err != nil {
		return err
	}
	decision, announced, released := g.close(callID, approve && remember)
	waitAnnounced(announced)
	if g.decided != nil {
		g.decided(callID, approve)
	}
	if decision != nil {
		decision <- approve
	}
	for _, id := range released {
		_ = g.decide(ctx, id, true, by, Edit{}, false)
	}
	return nil
}

// record stores the decision on a call that is still pending.
func (g *gate) record(ctx context.Context, callID string, approve bool, by string) error {
	status := store.ApprovalDenied
	if approve {
		status = store.ApprovalApproved
	}
	applied, err := g.store.DecideApproval(ctx, g.runID, callID, status, by)
	if err == nil && !applied {
		err = ErrNoPendingApproval
	}
	return err
}

// close drops a decided call from the gate. With grant set it remembers the
// approval on the call's scope and returns the other waiting calls it covers.
func (g *gate) close(callID string, grant bool) (chan bool, chan struct{}, []string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	decision := g.waiting[callID]
	name := g.edits.calls[callID].name
	scope := g.scopes[callID]
	delete(g.scopes, callID)
	var released []string
	if grant && scope != "" && g.takesGrant(name) {
		g.grants[name+"\x00"+scope] = true
		for id, call := range g.edits.calls {
			if id != callID && call.name == name && g.scopes[id] == scope && g.waiting[id] != nil {
				released = append(released, id)
			}
		}
	}
	announced := g.announced[callID]
	delete(g.announced, callID)
	delete(g.waiting, callID)
	delete(g.edits.calls, callID)
	return decision, announced, released
}

// waitAnnounced holds a decision until its call's pending event is out, so a
// decision never reaches the owner before the card it decides. It gives up
// after a few seconds rather than hold the run.
func waitAnnounced(announced chan struct{}) {
	if announced == nil {
		return
	}
	select {
	case <-announced:
	case <-time.After(5 * time.Second):
	}
}
