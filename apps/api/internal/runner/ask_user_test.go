package runner_test

import (
	"context"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/live"
	"github.com/FacileStudio/Bulle/apps/api/internal/push"
	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
)

const askCard = `{"questions":[{"question":"Which day?","options":[{"label":"Monday"},{"label":"Tuesday"}]}]}`

func TestAskUserYieldsTheTurnAndLeavesTheSessionWaiting(t *testing.T) {
	hub := push.NewHub(push.HubConfig{})
	mine, stop := hub.Subscribe(1)
	defer stop()
	backend := &fakeBackend{calls: []call{{id: "q1", name: "ask_user", input: askCard}}}
	f := newFixture(t, backend, nil, func(c *runner.Config) { c.Push = hub })
	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "plan my week")
	if err != nil {
		t.Fatal(err)
	}
	run := f.waitDone(t, runID)
	if run.StopReason != "yield" || run.Status != "finished" {
		t.Fatalf("run ended %s/%s, want finished/yield", run.Status, run.StopReason)
	}
	events := collectUntil(t, mine, live.RunFinished)
	if got := decode(t, events[len(events)-1])["session_status"]; got != live.StatusWaiting {
		t.Fatalf("session after a question: %v", got)
	}
	messages, _ := f.store.Messages(context.Background(), f.session.ID)
	for _, m := range messages {
		if _, text, _ := runner.MessageText(m.Payload); strings.Contains(text, "All done") || strings.Contains(text, "before I answered") {
			t.Fatalf("a yielded turn asked the model again or padded the transcript: %q", text)
		}
	}
}
