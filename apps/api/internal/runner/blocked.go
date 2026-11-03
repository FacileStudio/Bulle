package runner

import "time"

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
