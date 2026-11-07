package runner

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

const (
	deltaFlushAfter = 250 * time.Millisecond
	deltaFlushBytes = 4 << 10
)

// emitter stores a run's events and wakes its watchers. Text and thinking
// deltas are batched into one row per 250 ms or 4 KB, and any other event
// flushes the batch first, so the stored order stays the streamed order.
type emitter struct {
	store  store.Store
	hub    *hub
	logger *slog.Logger
	runID  string
	mu     sync.Mutex
	seq    int
	kind   string
	batch  strings.Builder
	timer  *time.Timer
}

func (e *emitter) emit(eventType string, payload any) {
	if delta, ok := deltaOf(eventType, payload); ok {
		e.buffer(eventType, delta)
		return
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		e.logger.Error("encoding run event", slog.String("type", eventType), slog.Any("error", err))
		return
	}
	e.mu.Lock()
	e.flushLocked()
	e.persistLocked(eventType, raw)
	e.mu.Unlock()
	e.hub.notify(e.runID)
}

func deltaOf(eventType string, payload any) (string, bool) {
	if eventType != EventText && eventType != EventThinking {
		return "", false
	}
	fields, ok := payload.(map[string]any)
	if !ok || len(fields) != 1 {
		return "", false
	}
	delta, ok := fields["delta"].(string)
	return delta, ok
}

func (e *emitter) buffer(kind, delta string) {
	e.mu.Lock()
	if e.kind != kind {
		e.flushLocked()
		e.kind = kind
	}
	e.batch.WriteString(delta)
	full := e.batch.Len() >= deltaFlushBytes
	if full {
		e.flushLocked()
	} else if e.timer == nil {
		e.timer = time.AfterFunc(deltaFlushAfter, e.flush)
	}
	e.mu.Unlock()
	if full {
		e.hub.notify(e.runID)
	}
}

func (e *emitter) flush() {
	e.mu.Lock()
	flushed := e.flushLocked()
	e.mu.Unlock()
	if flushed {
		e.hub.notify(e.runID)
	}
}

func (e *emitter) flushLocked() bool {
	if e.timer != nil {
		e.timer.Stop()
		e.timer = nil
	}
	if e.batch.Len() == 0 {
		return false
	}
	raw, _ := json.Marshal(map[string]string{"delta": e.batch.String()})
	e.batch.Reset()
	e.persistLocked(e.kind, raw)
	return true
}

func (e *emitter) persistLocked(eventType string, raw json.RawMessage) {
	e.seq++
	event := store.Event{RunID: e.runID, Seq: e.seq, Type: eventType, Payload: raw}
	if err := e.store.AppendEvent(context.Background(), event); err != nil {
		e.logger.Error("persisting run event", slog.String("run", e.runID), slog.String("type", eventType), slog.Any("error", err))
	}
}
