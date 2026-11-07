package runner

import (
	"context"
	"time"
)

// maxApprovalCredit caps how far waiting on the owner can push a run's
// deadline back, so a card nobody answers still frees the run slot.
const maxApprovalCredit = 2 * time.Hour

// deadline is context.WithTimeoutCause for a run whose budget counts only
// the time it worked: time spent waiting on an approval card, up to
// maxApprovalCredit, moves the deadline back. It ends ctx with errTimeout.
func (g *gate) deadline(parent context.Context, budget time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancelCause(parent)
	go g.expire(ctx, cancel, time.Now(), budget)
	return ctx, func() { cancel(nil) }
}

func (g *gate) expire(ctx context.Context, cancel context.CancelCauseFunc, start time.Time, budget time.Duration) {
	timer := time.NewTimer(budget)
	defer timer.Stop()
	floor := min(budget/10, time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
		left := budget + min(g.blockedSoFar(), maxApprovalCredit) - time.Since(start)
		if left <= 0 {
			cancel(errTimeout)
			return
		}
		timer.Reset(max(left, floor))
	}
}

// blockedSoFar is the time the run has waited on the owner, the wait in
// progress included.
func (g *gate) blockedSoFar() time.Duration {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.blocking == 0 {
		return g.blocked
	}
	return g.blocked + g.now().Sub(g.blockedAt)
}
