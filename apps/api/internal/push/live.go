package push

import (
	"encoding/json"
	"sync"
)

// EventNotification is the live event that carries a notification.
const EventNotification = "notification"

const liveBuffer = 64

// MaxStreams caps one user's open live streams. A new stream past it closes
// the oldest one, which is the tab most likely left behind.
const MaxStreams = 10

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
	mu     sync.Mutex
	subs   map[int64]map[chan Event]uint64
	next   uint64
	closed bool
}

// NewLive returns a Live with no listeners.
func NewLive() *Live {
	return &Live{subs: map[int64]map[chan Event]uint64{}}
}

// Subscribe returns the user's events and a function that stops them. The
// channel is closed when the listener falls too far behind, when the user
// opens more than MaxStreams, or when the server shuts down, so its stream
// ends and the client resyncs when it reconnects.
func (l *Live) Subscribe(userID int64) (<-chan Event, func()) {
	ch := make(chan Event, liveBuffer)
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		close(ch)
		return ch, func() {}
	}
	if l.subs[userID] == nil {
		l.subs[userID] = map[chan Event]uint64{}
	}
	if len(l.subs[userID]) >= MaxStreams {
		l.evictOldest(userID)
	}
	l.next++
	l.subs[userID][ch] = l.next
	return ch, func() {
		l.mu.Lock()
		defer l.mu.Unlock()
		if _, ok := l.subs[userID][ch]; ok {
			l.drop(userID, ch)
		}
	}
}

func (l *Live) evictOldest(userID int64) {
	var oldest chan Event
	var order uint64
	for ch, at := range l.subs[userID] {
		if oldest == nil || at < order {
			oldest, order = ch, at
		}
	}
	l.drop(userID, oldest)
	close(oldest)
}

// Close ends every open stream and refuses new ones. The server calls it
// when it starts shutting down, so no stream holds the drain open.
func (l *Live) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.closed = true
	for userID, subs := range l.subs {
		for ch := range subs {
			l.drop(userID, ch)
			close(ch)
		}
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
