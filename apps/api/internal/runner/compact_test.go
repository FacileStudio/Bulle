package runner_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"strings"
	"sync"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/core/compaction"
	"github.com/FacileStudio/nacelle"
)

type windowBackend struct {
	mu        sync.Mutex
	refuse    int
	summaries int
	mains     int
}

func (*windowBackend) Name() string { return "window" }

func (*windowBackend) Capabilities() nacelle.Capabilities {
	return nacelle.Capabilities{ContextWindow: 200_000, Thinking: true}
}

func (*windowBackend) CountTokens(context.Context, nacelle.Request) (int64, error) { return 0, nil }

func (w *windowBackend) Stream(_ context.Context, req nacelle.Request) iter.Seq2[nacelle.Event, error] {
	w.mu.Lock()
	summary := req.System == compaction.SystemPrompt
	refused := false
	if summary {
		w.summaries++
	} else {
		w.mains++
		refused = w.refuse > 0
		w.refuse--
	}
	w.mu.Unlock()
	return func(yield func(nacelle.Event, error) bool) {
		if refused {
			yield(nacelle.Event{}, errors.New("prompt is too long: 250000 tokens > 200000 maximum"))
			return
		}
		text := "All done."
		if summary {
			text = "Decisions:\n- read the whole repository"
		}
		if yield(nacelle.Event{Kind: nacelle.KindText, Text: text}, nil) {
			yield(nacelle.Event{Kind: nacelle.KindDone, Stop: nacelle.StopEnd}, nil)
		}
	}
}

func (w *windowBackend) counts() (summaries, mains int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.summaries, w.mains
}

func seedHistory(t *testing.T, f *fixture, turns int) {
	t.Helper()
	messages := []nacelle.Message{nacelle.UserText("the task")}
	for i := range turns {
		messages = append(messages, nacelle.AssistantText(fmt.Sprintf("step %d ", i)+strings.Repeat("a", 6_000)), nacelle.UserText(strings.Repeat("u", 6_000)))
	}
	messages = append(messages, nacelle.AssistantText("ready"))
	payloads := make([]json.RawMessage, len(messages))
	for i, message := range messages {
		raw, err := runner.EncodeMessage(message, nil)
		if err != nil {
			t.Fatal(err)
		}
		payloads[i] = raw
	}
	if err := f.store.ReplaceMessages(context.Background(), f.session.ID, payloads); err != nil {
		t.Fatal(err)
	}
}

func (f *fixture) ledgered(t *testing.T) bool {
	t.Helper()
	stored, err := f.store.Messages(context.Background(), f.session.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range stored {
		message, err := runner.DecodeMessage(row.Payload)
		if err != nil {
			t.Fatal(err)
		}
		if compaction.IsLedger(message) {
			return true
		}
	}
	return false
}

func TestALongHistoryIsCompactedBeforeTheRun(t *testing.T) {
	backend := &windowBackend{}
	f := newFixture(t, backend, nil, nil)
	seedHistory(t, f, 60)

	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "go on")
	if err != nil {
		t.Fatal(err)
	}
	if run := f.waitDone(t, runID); run.Status != "finished" {
		t.Fatalf("run status = %q, want finished: %s", run.Status, f.waitEvent(t, runID, runner.EventRunError).Payload)
	}
	if summaries, mains := backend.counts(); summaries != 1 || mains != 1 {
		t.Fatalf("summaries = %d, mains = %d, want one pass before one run", summaries, mains)
	}
	if !f.ledgered(t) {
		t.Error("the stored history carries no ledger after the pass")
	}
}

func TestAnOverflowForcesOnePassAndOneRetry(t *testing.T) {
	backend := &windowBackend{refuse: 1}
	f := newFixture(t, backend, nil, nil)
	seedHistory(t, f, 34)

	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "go on")
	if err != nil {
		t.Fatal(err)
	}
	if run := f.waitDone(t, runID); run.Status != "finished" {
		t.Fatalf("run status = %q, want finished after the retry", run.Status)
	}
	if summaries, mains := backend.counts(); summaries != 1 || mains != 2 {
		t.Fatalf("summaries = %d, mains = %d, want one forced pass and one retry", summaries, mains)
	}
	if !f.ledgered(t) {
		t.Error("the stored history carries no ledger after the forced pass")
	}
}
