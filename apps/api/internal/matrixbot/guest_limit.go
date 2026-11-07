package matrixbot

import (
	"sync"
	"time"
)

const (
	guestBurst  = 5
	guestWindow = 10 * time.Minute
)

// slowDown answers a colleague who wrote more than guestBurst messages to the
// owner's bot inside guestWindow.
const slowDown = "You have sent several messages in a short time. Try again in a few minutes."

type guestKey struct {
	owner  int64
	sender string
}

// guestLimiter caps how often one colleague can run the owner's model, so a
// guest cannot spend the owner's money or fill the owner's run slots.
type guestLimiter struct {
	mu   sync.Mutex
	seen map[guestKey][]time.Time
}

func (l *guestLimiter) allow(owner int64, sender string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.seen == nil {
		l.seen = map[guestKey][]time.Time{}
	}
	key := guestKey{owner: owner, sender: sender}
	var recent []time.Time
	for _, at := range l.seen[key] {
		if now.Sub(at) < guestWindow {
			recent = append(recent, at)
		}
	}
	if len(recent) >= guestBurst {
		l.seen[key] = recent
		return false
	}
	l.seen[key] = append(recent, now)
	return true
}
