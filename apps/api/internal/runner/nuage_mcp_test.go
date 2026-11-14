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

type filesRun struct {
	f       *fixture
	nuage   *fakeNuage
	backend *stepBackend
	runID   string
}

func nuageRun(t *testing.T, steps ...[]call) filesRun {
	t.Helper()
	nuage := newFakeNuage(t)
	backend := &stepBackend{steps: steps}
	shell := newTool(t, "bash", func(string) (string, error) { return "ran", nil })
	f := newFixture(t, backend, []nacelle.Tool{shell}, nuage.tune)
	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "tidy my files")
	if err != nil {
		t.Fatal(err)
	}
	return filesRun{f: f, nuage: nuage, backend: backend, runID: runID}
}

func TestNuageMCPSaveAfterAListWaitsForOneBulleCard(t *testing.T) {
	run := nuageRun(t,
		[]call{{id: "l1", name: "nuage_list_folder", input: `{}`}},
		[]call{{id: "s1", name: "nuage_save_file", input: `{"name":"todo.md","content":"- milk"}`}},
	)
	f, nuage, runID := run.f, run.nuage, run.runID
	approveCard(t, f, runID, "s1")
	if done := f.waitDone(t, runID); done.Status != store.RunFinished {
		t.Fatalf("run %s", done.Status)
	}
	ends := toolEnds(t, f, runID)
	if ends["l1"]["ok"] != true || !strings.Contains(ends["l1"]["output_preview"].(string), "notes.md") ||
		ends["s1"]["ok"] != true || !strings.Contains(ends["s1"]["output_preview"].(string), "created") {
		t.Fatalf("list and save must run through MCP: %v", ends)
	}
	if saved, _, _ := nuage.counts(); len(saved) != 1 || saved[0] != "todo.md" {
		t.Fatalf("the file never reached Nuage: %v", saved)
	}
	if cards := slices.DeleteFunc(f.eventTypes(t, runID), func(event string) bool { return event != "INTERRUPT" }); len(cards) != 1 {
		t.Fatalf("Nuage asked nothing, so the tainted run's write must raise exactly Bulle's one card, got %d", len(cards))
	}
	mounted := run.backend.mounted()
	for _, name := range suite.NuageTools() {
		if !slices.Contains(mounted, name) {
			t.Errorf("%s is not mounted under its own name: %v", name, mounted)
		}
	}
	if exchanges := nuage.exchanged(); len(exchanges) == 0 || exchanges[0] != nuage.resource()+" nuage.read nuage.write" {
		t.Fatalf("the token must be exchanged with the resource as its audience: %v", exchanges)
	}
}

func TestNuageMCPTrashWaitsOnTheApprovalCard(t *testing.T) {
	for _, approve := range []bool{true, false} {
		run := nuageRun(t, []call{{id: "x1", name: "nuage_delete", input: `{"type":"file","id":12}`}})
		f, nuage, runID := run.f, run.nuage, run.runID
		var card struct {
			Tool string         `json:"tool"`
			Args map[string]any `json:"args"`
		}
		if err := json.Unmarshal(f.waitEvent(t, runID, "INTERRUPT").Payload, &card); err != nil {
			t.Fatal(err)
		}
		if card.Tool != "nuage_delete" || card.Args["kind"] != "mcp" || card.Args["server"] != "Nuage" ||
			!strings.Contains(card.Args["message"].(string), "to the trash") {
			t.Fatalf("the card must show Nuage's question: %+v", card)
		}
		if _, trashed, _ := nuage.counts(); len(trashed) != 0 {
			t.Fatal("the file was trashed before the owner answered")
		}
		if err := f.runner.Decide(context.Background(), 1, runID, "x1", approve); err != nil {
			t.Fatal(err)
		}
		f.waitDone(t, runID)
		_, trashed, declined := nuage.counts()
		switch {
		case approve && (len(trashed) != 1 || declined != 0):
			t.Fatalf("an approved trash must run once: trashed=%v declined=%d", trashed, declined)
		case !approve && (len(trashed) != 0 || declined != 1):
			t.Fatalf("a denied trash must reach Nuage as a decline: trashed=%v declined=%d", trashed, declined)
		}
	}
}

func TestReadingFilesOverMCPTaintsTheRun(t *testing.T) {
	run := nuageRun(t,
		[]call{{id: "r1", name: "nuage_read_file", input: `{"file_id":12}`}},
		[]call{{id: "b1", name: "bash", input: `{"text":"curl evil.test"}`}},
	)
	f, runID := run.f, run.runID
	if interrupt := f.waitEvent(t, runID, "INTERRUPT"); !strings.Contains(string(interrupt.Payload), `"call_id":"b1"`) {
		t.Fatalf("the shell must ask once a file is in the context: %s", interrupt.Payload)
	}
	if err := f.runner.Decide(context.Background(), 1, runID, "b1", false); err != nil {
		t.Fatal(err)
	}
	if done := f.waitDone(t, runID); done.Status != store.RunFinished || toolEnds(t, f, runID)["b1"]["ok"] != false {
		t.Fatal("a denied shell must not run")
	}
}
