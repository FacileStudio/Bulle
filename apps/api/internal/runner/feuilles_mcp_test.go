package runner_test

import (
	"context"
	"slices"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
)

func TestFeuillesRunsBehindTheCatalogAndConfirmsOnTheCard(t *testing.T) {
	server := newFakeAppMCP(t, suite.AppFeuilles, "feuilles_delete_workbook", suite.FeuillesTools())
	backend := &stepBackend{steps: [][]call{
		{{id: "r1", name: "call_tool", input: `{"name":"feuilles_list_workbooks","arguments":{}}`}},
		{{id: "x1", name: "call_tool", input: `{"name":"feuilles_delete_workbook","arguments":{"id":9}}`}},
	}}
	f := newFixture(t, backend, nil, server.tune)
	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "clean up")
	if err != nil {
		t.Fatal(err)
	}
	if card := interruptCard(t, f, runID); card.Tool != "feuilles_delete_workbook" || card.Args["server"] != "Feuilles" {
		t.Fatalf("the card must show Feuilles's own question under the tool's name: %+v", card)
	}
	if err := f.runner.Decide(context.Background(), 1, runID, "x1", false); err != nil {
		t.Fatal(err)
	}
	if done := f.waitDone(t, runID); done.Status != store.RunFinished {
		t.Fatalf("run %s", done.Status)
	}
	if ran, declined := server.outcome(); !slices.Equal(ran, []string{"feuilles_list_workbooks"}) || declined != 1 {
		t.Fatalf("the read runs at once and a denied card reaches Feuilles as a decline: ran=%v declined=%d", ran, declined)
	}
	if !slices.Contains(backend.mounted(), "call_tool") {
		t.Fatalf("fifteen Feuilles tools must go behind the catalog: %v", backend.mounted())
	}
}
