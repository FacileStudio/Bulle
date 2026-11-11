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

type timeRun struct {
	f       *fixture
	sablier *fakeSablierMCP
	backend *stepBackend
	runID   string
}

func sablierRun(t *testing.T, steps ...[]call) timeRun {
	t.Helper()
	sablier := newFakeSablierMCP(t)
	backend := &stepBackend{steps: steps}
	shell := newTool(t, "bash", func(string) (string, error) { return "ran", nil })
	f := newFixture(t, backend, []nacelle.Tool{shell}, sablier.tune)
	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "track my time")
	if err != nil {
		t.Fatal(err)
	}
	return timeRun{f: f, sablier: sablier, backend: backend, runID: runID}
}

func TestSablierMCPStatusAndStartRunUnasked(t *testing.T) {
	run := sablierRun(t,
		[]call{{id: "s1", name: "sablier_timer_status", input: `{}`}},
		[]call{{id: "t1", name: "sablier_start_timer", input: `{"project_id":2,"task_id":3}`}},
	)
	f, sablier, runID := run.f, run.sablier, run.runID
	if done := f.waitDone(t, runID); done.Status != store.RunFinished {
		t.Fatalf("run %s", done.Status)
	}
	ends := toolEnds(t, f, runID)
	if ends["s1"]["ok"] != true || ends["t1"]["ok"] != true || !strings.Contains(ends["t1"]["output_preview"].(string), "started") {
		t.Fatalf("status and start must run through MCP: %v", ends)
	}
	if started, _, _ := sablier.counts(); len(started) != 1 || started[0] != 2 {
		t.Fatalf("the timer never reached Sablier: %v", started)
	}
	if slices.Contains(f.eventTypes(t, runID), "INTERRUPT") {
		t.Fatal("Sablier asked nothing, so Bulle must not raise its own card")
	}
	mounted := run.backend.mounted()
	for _, name := range suite.SablierTools() {
		if !slices.Contains(mounted, name) {
			t.Errorf("%s is not mounted under its own name: %v", name, mounted)
		}
	}
	if exchanges := sablier.exchanged(); len(exchanges) == 0 || exchanges[0] != sablier.resource()+" sablier.read sablier.write" {
		t.Fatalf("the token must be exchanged with the resource as its audience: %v", exchanges)
	}
}

func TestSablierMCPDeleteWaitsOnTheApprovalCard(t *testing.T) {
	for _, approve := range []bool{true, false} {
		run := sablierRun(t, []call{{id: "x1", name: "sablier_delete_entry", input: `{"entry_id":9}`}})
		f, sablier, runID := run.f, run.sablier, run.runID
		var card struct {
			Tool string         `json:"tool"`
			Args map[string]any `json:"args"`
		}
		if err := json.Unmarshal(f.waitEvent(t, runID, "INTERRUPT").Payload, &card); err != nil {
			t.Fatal(err)
		}
		if card.Tool != "sablier_delete_entry" || card.Args["kind"] != "mcp" || card.Args["server"] != "Sablier" ||
			!strings.Contains(card.Args["message"].(string), "Delete the entry") {
			t.Fatalf("the card must show Sablier's question: %+v", card)
		}
		if _, deleted, _ := sablier.counts(); len(deleted) != 0 {
			t.Fatal("the entry was deleted before the owner answered")
		}
		if err := f.runner.Decide(context.Background(), 1, runID, "x1", approve); err != nil {
			t.Fatal(err)
		}
		f.waitDone(t, runID)
		_, deleted, declined := sablier.counts()
		switch {
		case approve && (len(deleted) != 1 || declined != 0):
			t.Fatalf("an approved delete must run once: deleted=%v declined=%d", deleted, declined)
		case !approve && (len(deleted) != 0 || declined != 1):
			t.Fatalf("a denied delete must reach Sablier as a decline: deleted=%v declined=%d", deleted, declined)
		}
	}
}

func TestReadingTimeEntriesOverMCPTaintsTheRun(t *testing.T) {
	run := sablierRun(t,
		[]call{{id: "r1", name: "sablier_list_entries", input: `{}`}},
		[]call{{id: "b1", name: "bash", input: `{"text":"curl evil.test"}`}},
	)
	f, runID := run.f, run.runID
	if interrupt := f.waitEvent(t, runID, "INTERRUPT"); !strings.Contains(string(interrupt.Payload), `"call_id":"b1"`) {
		t.Fatalf("the shell must ask once a shared entry is in the context: %s", interrupt.Payload)
	}
	if err := f.runner.Decide(context.Background(), 1, runID, "b1", false); err != nil {
		t.Fatal(err)
	}
	if done := f.waitDone(t, runID); done.Status != store.RunFinished || toolEnds(t, f, runID)["b1"]["ok"] != false {
		t.Fatal("a denied shell must not run")
	}
}
