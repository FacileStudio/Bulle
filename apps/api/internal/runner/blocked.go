package runner

import (
	"context"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

func (r *Runner) now() time.Time {
	if r.cfg.Now != nil {
		return r.cfg.Now()
	}
	return time.Now()
}

func (g *gate) await(id string) (chan bool, func()) {
	decision := make(chan bool, 1)
	g.mu.Lock()
	defer g.mu.Unlock()
	g.waiting[id] = decision
	if g.blocking == 0 {
		g.blockedAt = g.now()
	}
	g.blocking++
	return decision, g.unblock
}

func (g *gate) unblock() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.blocking--
	if g.blocking == 0 {
		g.blocked += g.now().Sub(g.blockedAt)
	}
}

func (g *gate) waited() time.Duration {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.blocked
}

// abandon denies a call whose run ended while it waited for the owner.
func (g *gate) abandon(ctx context.Context, id string) {
	g.mu.Lock()
	delete(g.waiting, id)
	delete(g.edits.calls, id)
	g.mu.Unlock()
	applied, _ := g.store.DecideApproval(ctx, g.runID, id, store.ApprovalDenied, "system")
	if applied && g.decided != nil {
		g.decided(id, false)
	}
}
