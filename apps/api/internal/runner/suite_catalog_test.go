package runner_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

func TestASuiteToolBehindTheCatalogAsksUnderItsOwnName(t *testing.T) {
	glouton := newFakeGloutonMCP(t)
	backend := &fakeBackend{calls: []call{{id: "o1", name: "call_tool",
		input: `{"name":"glouton_company_delete","arguments":{"company_id":"c1"}}`}}}
	f := newFixture(t, backend, nil, glouton.tune)
	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "add a task")
	if err != nil {
		t.Fatal(err)
	}
	var card struct {
		CallID string         `json:"call_id"`
		Tool   string         `json:"tool"`
		Args   map[string]any `json:"args"`
	}
	if err := json.Unmarshal(f.waitEvent(t, runID, "INTERRUPT").Payload, &card); err != nil {
		t.Fatal(err)
	}
	if card.CallID != "o1" || card.Tool != "glouton_company_delete" || card.Args["server"] != "Glouton" {
		t.Fatalf("the approval card must name the app tool and its arguments: %+v", card)
	}
	var started struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(f.waitEvent(t, runID, "TOOL_CALL_START").Payload, &started); err != nil {
		t.Fatal(err)
	}
	if started.Name != "glouton_company_delete" {
		t.Fatalf("the tool line must show the app tool, got %q", started.Name)
	}
	if err := f.runner.Decide(context.Background(), 1, runID, "o1", false); err != nil {
		t.Fatal(err)
	}
	if run := f.waitDone(t, runID); run.Status != store.RunFinished {
		t.Fatalf("run %s", run.Status)
	}
}
