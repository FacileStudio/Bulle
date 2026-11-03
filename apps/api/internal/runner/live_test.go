package runner_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/live"
	"github.com/FacileStudio/Bulle/apps/api/internal/push"
	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/nacelle"
)

func (r *recordedPush) Publish(int64, string, any) {}

func collectUntil(t *testing.T, events <-chan push.Event, last string) []push.Event {
	t.Helper()
	var out []push.Event
	timeout := time.After(5 * time.Second)
	for {
		select {
		case event := <-events:
			if event.Type == push.EventNotification {
				continue
			}
			out = append(out, event)
			if event.Type == last {
				return out
			}
		case <-timeout:
			t.Fatalf("no %s event, got %d events", last, len(out))
		}
	}
}

func decode(t *testing.T, event push.Event) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(event.Data, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func checkRunEvents(t *testing.T, got []push.Event, sessionID, runID string) {
	t.Helper()
	want := []string{live.RunStarted, live.SessionUpdated, live.ApprovalPending, live.ApprovalDecided, live.RunFinished}
	if len(got) != len(want) {
		t.Fatalf("events %d, want %v", len(got), want)
	}
	for i, event := range got {
		if data := decode(t, event); event.Type != want[i] || data["session_id"] != sessionID {
			t.Fatalf("event %d is %s for %v, want %s", i, event.Type, data["session_id"], want[i])
		}
	}
	checks := []struct {
		index int
		key   string
		want  any
	}{
		{1, "title", "mail bob"}, {1, "status", live.StatusRunning},
		{2, "call_id", "m1"}, {2, "session_status", live.StatusWaiting},
		{3, "approved", true}, {3, "session_status", live.StatusRunning},
		{4, "run_id", runID}, {4, "status", "finished"}, {4, "session_status", live.StatusIdle},
	}
	for _, c := range checks {
		if got := decode(t, got[c.index])[c.key]; got != c.want {
			t.Fatalf("%s %s = %v, want %v", want[c.index], c.key, got, c.want)
		}
	}
}

func TestLiveEventsFollowARunAndReachOnlyItsOwner(t *testing.T) {
	hub := push.NewHub(push.HubConfig{})
	mine, stop := hub.Subscribe(1)
	defer stop()
	other, stopOther := hub.Subscribe(2)
	defer stopOther()
	send := newTool(t, "send_mail", func(string) (string, error) { return "sent", nil })
	backend := &fakeBackend{calls: []call{{id: "m1", name: "send_mail", input: `{"text":"to bob"}`}}}
	f := newFixture(t, backend, []nacelle.Tool{send}, func(c *runner.Config) {
		c.Outward = []string{"send_mail"}
		c.Push = hub
	})

	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "mail bob")
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, "waiting status", func() bool { return f.runner.Statuses(1)[f.session.ID] == live.StatusWaiting })
	if err := f.runner.Decide(context.Background(), 1, runID, "m1", true); err != nil {
		t.Fatal(err)
	}
	checkRunEvents(t, collectUntil(t, mine, live.RunFinished), f.session.ID, runID)
	if final := collectUntil(t, mine, live.SessionUpdated); decode(t, final[0])["status"] != live.StatusIdle {
		t.Fatalf("session after the run: %s", final[0].Data)
	}
	if len(f.runner.Statuses(1)) != 0 || len(other) != 0 {
		t.Fatalf("statuses after the run %v, events for another user %d", f.runner.Statuses(1), len(other))
	}
}

func TestStatusIsRunningWhileATurnStreams(t *testing.T) {
	backend := &fakeBackend{block: make(chan struct{})}
	f := newFixture(t, backend, nil, nil)
	if len(f.runner.Statuses(1)) != 0 {
		t.Fatal("a session with no run should be idle")
	}
	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "hello")
	if err != nil {
		t.Fatal(err)
	}
	if got := f.runner.Statuses(1)[f.session.ID]; got != live.StatusRunning {
		t.Fatalf("status while streaming: %q", got)
	}
	if len(f.runner.Statuses(2)) != 0 {
		t.Fatal("another user sees this run")
	}
	close(backend.block)
	f.waitDone(t, runID)
	if len(f.runner.Statuses(1)) != 0 {
		t.Fatal("status after the run should be idle")
	}
}
