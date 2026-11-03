package push

import (
	"encoding/json"
	"sync"
)

// EventNotification is the live event that carries a notification.
const EventNotification = "notification"

const liveBuffer = 64

// Event is one message on a user's live stream: a notification, or a change
// to the user's sessions, runs and approvals.
type Event struct {
	Type string
	Data json.RawMessage
}

type liveNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Path  string `json:"path"`
}

// Live fans events out to the user's open apps, which listen on a
// server-sent event stream. One API replica is assumed.
type Live struct {
	mu   sync.Mutex
	subs map[int64]map[chan Event]struct{}
}

// NewLive returns a Live with no listeners.
func NewLive() *Live {
	return &Live{subs: map[int64]map[chan Event]struct{}{}}
}

// Subscribe returns the user's events and a function that stops them. The
// channel is closed when the listener falls too far behind, so its stream
// ends and the client resyncs when it reconnects.
func (l *Live) Subscribe(userID int64) (<-chan Event, func()) {
	ch := make(chan Event, liveBuffer)
	l.mu.Lock()
	if l.subs[userID] == nil {
		l.subs[userID] = map[chan Event]struct{}{}
	}
	l.subs[userID][ch] = struct{}{}
	l.mu.Unlock()
	return ch, func() {
		l.mu.Lock()
		defer l.mu.Unlock()
		l.drop(userID, ch)
	}
}

// Publish hands an event to every listener of the user and never blocks: a
// listener whose buffer is full is cut off instead.
func (l *Live) Publish(userID int64, eventType string, data any) {
	raw, err := json.Marshal(data)
	if err != nil {
		return
	}
	event := Event{Type: eventType, Data: raw}
	l.mu.Lock()
	defer l.mu.Unlock()
	for ch := range l.subs[userID] {
		select {
		case ch <- event:
		default:
			l.drop(userID, ch)
			close(ch)
		}
	}
}

func (l *Live) drop(userID int64, ch chan Event) {
	delete(l.subs[userID], ch)
	if len(l.subs[userID]) == 0 {
		delete(l.subs, userID)
	}
}

// Subscribe streams the user's live events to an open app.
func (h *Hub) Subscribe(userID int64) (<-chan Event, func()) {
	return h.cfg.Live.Subscribe(userID)
}

// Publish sends one live event to the user's open apps. A nil Hub drops it.
func (h *Hub) Publish(userID int64, eventType string, data any) {
	if h == nil {
		return
	}
	h.cfg.Live.Publish(userID, eventType, data)
}
