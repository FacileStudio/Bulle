package runner

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

func TestDeltasAreBatchedInStreamOrder(t *testing.T) {
	st := store.NewMemory()
	em := &emitter{store: st, hub: newHub(), logger: slog.Default(), runID: "r1"}
	for _, delta := range []string{"Hel", "lo ", "there"} {
		em.emit(EventText, map[string]any{"delta": delta})
	}
	em.emit(EventThinking, map[string]any{"delta": "hm"})
	em.emit(EventToolStart, map[string]any{"call_id": "c1"})
	em.emit(EventText, map[string]any{"delta": strings.Repeat("x", deltaFlushBytes)})
	events, _ := st.Events(context.Background(), "r1", 0)
	types := make([]string, len(events))
	for i, event := range events {
		types[i] = event.Type
	}
	want := []string{EventText, EventThinking, EventToolStart, EventText}
	if strings.Join(types, ",") != strings.Join(want, ",") || !strings.Contains(string(events[0].Payload), `"Hello there"`) {
		t.Fatalf("events %v, first %s", types, events[0].Payload)
	}
}
