package push

import "sync"

// Live fans notifications out to the user's open desktop apps, which listen
// on a server-sent event stream instead of a push service.
type Live struct {
	mu   sync.Mutex
	subs map[int64]map[chan Notification]struct{}
}

// NewLive returns a Live with no listeners.
func NewLive() *Live {
	return &Live{subs: map[int64]map[chan Notification]struct{}{}}
}

// Subscribe returns the user's notifications and a function that stops them.
func (l *Live) Subscribe(userID int64) (<-chan Notification, func()) {
	ch := make(chan Notification, 8)
	l.mu.Lock()
	if l.subs[userID] == nil {
		l.subs[userID] = map[chan Notification]struct{}{}
	}
	l.subs[userID][ch] = struct{}{}
	l.mu.Unlock()
	return ch, func() {
		l.mu.Lock()
		defer l.mu.Unlock()
		delete(l.subs[userID], ch)
		if len(l.subs[userID]) == 0 {
			delete(l.subs, userID)
		}
	}
}

// publish hands n to every listener of the user, dropping it for a listener
// whose buffer is full rather than blocking the sender.
func (l *Live) publish(userID int64, n Notification) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for ch := range l.subs[userID] {
		select {
		case ch <- n:
		default:
		}
	}
}
