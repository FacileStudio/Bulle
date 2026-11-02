package runner

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"unicode/utf8"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

const (
	EventRunStarted  = "RUN_STARTED"
	EventText        = "TEXT_MESSAGE_CONTENT"
	EventThinking    = "THINKING"
	EventToolStart   = "TOOL_CALL_START"
	EventToolEnd     = "TOOL_CALL_END"
	EventInterrupt   = "INTERRUPT"
	EventRunFinished = "RUN_FINISHED"
	EventRunError    = "RUN_ERROR"
)

func IsTerminal(eventType string) bool {
	return eventType == EventRunFinished || eventType == EventRunError
}

const previewLimit = 2000

type emitter struct {
	store  store.Store
	hub    *hub
	logger *slog.Logger
	runID  string
	mu     sync.Mutex
	seq    int
}

func (e *emitter) emit(eventType string, payload any) {
	raw, err := json.Marshal(payload)
	if err != nil {
		e.logger.Error("encoding run event", slog.String("type", eventType), slog.Any("error", err))
		return
	}
	e.mu.Lock()
	e.seq++
	event := store.Event{RunID: e.runID, Seq: e.seq, Type: eventType, Payload: raw}
	err = e.store.AppendEvent(context.Background(), event)
	e.mu.Unlock()
	if err != nil {
		e.logger.Error("persisting run event", slog.String("run", e.runID), slog.String("type", eventType), slog.Any("error", err))
	}
	e.hub.notify(e.runID)
}

func Preview(text string) string {
	if len(text) <= previewLimit {
		return text
	}
	cut := previewLimit
	for cut > 0 && !utf8.RuneStart(text[cut]) {
		cut--
	}
	return text[:cut] + "..."
}

func preview(text string) string {
	return Preview(text)
}

func rawArgs(input string) json.RawMessage {
	if json.Valid([]byte(input)) {
		return json.RawMessage(input)
	}
	quoted, _ := json.Marshal(input)
	return quoted
}

type hub struct {
	mu   sync.Mutex
	subs map[string]map[chan struct{}]struct{}
	seen map[string]int
}

func newHub() *hub {
	return &hub{subs: map[string]map[chan struct{}]struct{}{}, seen: map[string]int{}}
}

func (h *hub) watchers(runID string, delta int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.seen[runID] += delta
	if h.seen[runID] <= 0 {
		delete(h.seen, runID)
	}
}

func (h *hub) watched(runID string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.seen[runID] > 0
}

func (h *hub) subscribe(runID string) (<-chan struct{}, func()) {
	ch := make(chan struct{}, 1)
	h.mu.Lock()
	if h.subs[runID] == nil {
		h.subs[runID] = map[chan struct{}]struct{}{}
	}
	h.subs[runID][ch] = struct{}{}
	h.mu.Unlock()
	return ch, func() {
		h.mu.Lock()
		delete(h.subs[runID], ch)
		if len(h.subs[runID]) == 0 {
			delete(h.subs, runID)
		}
		h.mu.Unlock()
	}
}

func (h *hub) notify(runID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subs[runID] {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
