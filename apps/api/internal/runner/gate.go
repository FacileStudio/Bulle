package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

type announced struct {
	id    string
	input string
}

type gate struct {
	outward map[string]bool
	asks    map[string]func(input json.RawMessage) bool
	store   store.Store
	emitter *emitter
	runID   string
	pended  func(callID, tool string)
	decided func(callID string, approved bool)
	now     func() time.Time

	mu        sync.Mutex
	blocking  int
	blockedAt time.Time
	blocked   time.Duration
	asked     map[string][]announced
	waiting   map[string]chan bool
	orphans   int
	policies  map[string]approvalPolicy
	grants    map[string]bool
	edits     editState
}

func newGate(outward map[string]bool, st store.Store, em *emitter, runID string, now func() time.Time) *gate {
	return &gate{
		now:     now,
		outward: outward,
		store:   st,
		emitter: em,
		runID:   runID,
		asked:   map[string][]announced{},
		waiting: map[string]chan bool{},
		grants:  map[string]bool{},
	}
}

func normalize(input string) string {
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(input)); err != nil {
		return input
	}
	return buf.String()
}

func (g *gate) note(id, name, input string) {
	if !g.outward[name] && g.asks[name] == nil {
		return
	}
	g.mu.Lock()
	g.asked[name] = append(g.asked[name], announced{id: id, input: normalize(input)})
	g.mu.Unlock()
}

func (g *gate) claim(name, input string) (string, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	queue := g.asked[name]
	if len(queue) == 0 {
		return "", false
	}
	pick := 0
	want := normalize(input)
	for i, candidate := range queue {
		if candidate.input == want {
			pick = i
			break
		}
	}
	id := queue[pick].id
	g.asked[name] = append(queue[:pick:pick], queue[pick+1:]...)
	return id, true
}

func (g *gate) callID(name, input string) string {
	for range 50 {
		if id, ok := g.claim(name, input); ok {
			return id
		}
		time.Sleep(10 * time.Millisecond)
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.orphans++
	return fmt.Sprintf("unannounced-%d", g.orphans)
}

func (g *gate) approve(ctx context.Context, name string, input json.RawMessage) bool {
	if !g.outward[name] && g.asks[name] == nil {
		return true
	}
	id := g.callID(name, string(input))
	if !g.outward[name] && !g.asks[name](input) {
		return true
	}
	args, scope, ok := g.describe(ctx, name, input)
	if !ok || g.granted(name, scope) {
		return ok
	}
	background := context.WithoutCancel(ctx)
	if err := g.store.CreateApproval(background, g.runID, id); err != nil {
		return false
	}
	g.remember(id, name, input)
	decision, release := g.await(id)
	defer release()
	g.emitter.emit(EventInterrupt, map[string]any{"call_id": id, "tool": name, "args": args})
	if g.pended != nil {
		g.pended(id, name)
	}
	select {
	case approved := <-decision:
		g.settle(background, name, input, scope, approved)
		return approved
	case <-ctx.Done():
		g.abandon(background, id)
		return false
	}
}

func (g *gate) pending(callID string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	_, ok := g.waiting[callID]
	return ok
}

func (g *gate) decide(ctx context.Context, callID string, approve bool, by string, edit Edit) error {
	if !g.pending(callID) {
		return ErrNoPendingApproval
	}
	if approve && !edit.empty() {
		if err := g.revise(callID, edit); err != nil {
			return err
		}
	}
	status := store.ApprovalDenied
	if approve {
		status = store.ApprovalApproved
	}
	applied, err := g.store.DecideApproval(ctx, g.runID, callID, status, by)
	if err != nil {
		return err
	}
	if !applied {
		return ErrNoPendingApproval
	}
	g.mu.Lock()
	decision := g.waiting[callID]
	delete(g.waiting, callID)
	delete(g.edits.calls, callID)
	g.mu.Unlock()
	if decision != nil {
		decision <- approve
	}
	if g.decided != nil {
		g.decided(callID, approve)
	}
	return nil
}
