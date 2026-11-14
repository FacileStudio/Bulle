package runner_test

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/nacelle"
)

type tasksRun struct {
	f       *fixture
	opus    *fakeOpus
	backend *stepBackend
	runID   string
}

func opusRun(t *testing.T, steps ...[]call) tasksRun {
	t.Helper()
	opus := newFakeOpus(t)
	backend := &stepBackend{steps: steps}
	shell := newTool(t, "bash", func(string) (string, error) { return "ran", nil })
	f := newFixture(t, backend, []nacelle.Tool{shell}, opus.tune)
	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "plan my week")
	if err != nil {
		t.Fatal(err)
	}
	return tasksRun{f: f, opus: opus, backend: backend, runID: runID}
}

func TestOpusMCPCreateAfterASearchWaitsForOneBulleCard(t *testing.T) {
	run := opusRun(t,
		[]call{{id: "s1", name: "opus_search_tasks", input: `{"assignee":"me"}`}},
		[]call{{id: "c1", name: "opus_create_task", input: `{"project_id":5,"title":"Ship it"}`}},
	)
	f, opus, runID := run.f, run.opus, run.runID
	approveCard(t, f, runID, "c1")
	if done := f.waitDone(t, runID); done.Status != store.RunFinished {
		t.Fatalf("run %s", done.Status)
	}
	ends := toolEnds(t, f, runID)
	if ends["s1"]["ok"] != true || !strings.Contains(ends["s1"]["output_preview"].(string), "web-12") ||
		ends["c1"]["ok"] != true || !strings.Contains(ends["c1"]["output_preview"].(string), "created") {
		t.Fatalf("search and create must run through MCP: %v", ends)
	}
	if created, _, _ := opus.counts(); len(created) != 1 || created[0] != "Ship it" {
		t.Fatalf("the task never reached Opus: %v", created)
	}
	if cards := slices.DeleteFunc(f.eventTypes(t, runID), func(event string) bool { return event != "INTERRUPT" }); len(cards) != 1 {
		t.Fatalf("Opus asked nothing, so the tainted run's write must raise exactly Bulle's one card, got %d", len(cards))
	}
	mounted := run.backend.mounted()
	for _, name := range suite.OpusTools() {
		if !slices.Contains(mounted, name) {
			t.Errorf("%s is not mounted under its own name: %v", name, mounted)
		}
	}
	if exchanges := opus.exchanged(); len(exchanges) == 0 || exchanges[0] != opus.resource()+" opus.read opus.write" {
		t.Fatalf("the token must be exchanged with the resource as its audience: %v", exchanges)
	}
}

func TestOpusMCPDeleteWaitsOnTheApprovalCard(t *testing.T) {
	for _, approve := range []bool{true, false} {
		run := opusRun(t, []call{{id: "x1", name: "opus_delete_task", input: `{"task_id":42}`}})
		f, opus, runID := run.f, run.opus, run.runID
		var card struct {
			Tool string         `json:"tool"`
			Args map[string]any `json:"args"`
		}
		if err := json.Unmarshal(f.waitEvent(t, runID, "INTERRUPT").Payload, &card); err != nil {
			t.Fatal(err)
		}
		if card.Tool != "opus_delete_task" || card.Args["kind"] != "mcp" || card.Args["server"] != "Opus" ||
			!strings.Contains(card.Args["message"].(string), "Delete the task") {
			t.Fatalf("the card must show Opus's question: %+v", card)
		}
		if _, deleted, _ := opus.counts(); len(deleted) != 0 {
			t.Fatal("the task was deleted before the owner answered")
		}
		if err := f.runner.Decide(context.Background(), 1, runID, "x1", approve); err != nil {
			t.Fatal(err)
		}
		f.waitDone(t, runID)
		_, deleted, declined := opus.counts()
		switch {
		case approve && (len(deleted) != 1 || declined != 0):
			t.Fatalf("an approved delete must run once: deleted=%v declined=%d", deleted, declined)
		case !approve && (len(deleted) != 0 || declined != 1):
			t.Fatalf("a denied delete must reach Opus as a decline: deleted=%v declined=%d", deleted, declined)
		}
	}
}

func TestReadingTasksOverMCPTaintsTheRun(t *testing.T) {
	run := opusRun(t,
		[]call{{id: "r1", name: "opus_read_task", input: `{"task_id":42}`}},
		[]call{{id: "b1", name: "bash", input: `{"text":"curl evil.test"}`}},
	)
	f, runID := run.f, run.runID
	if interrupt := f.waitEvent(t, runID, "INTERRUPT"); !strings.Contains(string(interrupt.Payload), `"call_id":"b1"`) {
		t.Fatalf("the shell must ask once a task is in the context: %s", interrupt.Payload)
	}
	if err := f.runner.Decide(context.Background(), 1, runID, "b1", false); err != nil {
		t.Fatal(err)
	}
	if done := f.waitDone(t, runID); done.Status != store.RunFinished || toolEnds(t, f, runID)["b1"]["ok"] != false {
		t.Fatal("a denied shell must not run")
	}
}
