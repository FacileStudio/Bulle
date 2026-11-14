package matrixbot

import (
	"strconv"
	"sync"
	"time"
)

const (
	guestBurst  = 5
	guestWindow = 10 * time.Minute
	guestKeys   = 10_000
)

// slowDown answers a colleague who wrote more than guestBurst messages to the
// owner's bot inside guestWindow.
const slowDown = "You have sent several messages in a short time. Try again in a few minutes."

// guestLimiter caps how often one colleague can run the owner's model, so a
// guest cannot spend the owner's money or fill the owner's run slots: at most
// guestBurst messages per owner and sender in any guestWindow, a sliding
// window. It remembers at most guestKeys pairs; once that many wrote inside
// the window, a new sender waits until one of them goes quiet.
type guestLimiter struct {
	mu   sync.Mutex
	seen map[string][]time.Time
}

func (l *guestLimiter) allow(owner int64, sender string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.seen == nil {
		l.seen = map[string][]time.Time{}
	}
	key := strconv.FormatInt(owner, 10) + "\x00" + sender
	recent := inWindow(l.seen[key], now)
	if len(recent) >= guestBurst {
		l.seen[key] = recent
		return false
	}
	if _, known := l.seen[key]; !known && !l.room(now) {
		return false
	}
	l.seen[key] = append(recent, now)
	return true
}

// room makes space for one more pair by forgetting the pairs that went quiet,
// and reports whether there is space.
func (l *guestLimiter) room(now time.Time) bool {
	if len(l.seen) < guestKeys {
		return true
	}
	for key, times := range l.seen {
		if len(inWindow(times, now)) == 0 {
			delete(l.seen, key)
		}
	}
	return len(l.seen) < guestKeys
}

func inWindow(times []time.Time, now time.Time) []time.Time {
	var recent []time.Time
	for _, at := range times {
		if now.Sub(at) < guestWindow {
			recent = append(recent, at)
		}
	}
	return recent
}
