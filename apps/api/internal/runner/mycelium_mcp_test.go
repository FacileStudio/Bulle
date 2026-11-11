package runner_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestMemoryDeleteWaitsOnMyceliumsCard(t *testing.T) {
	for _, approve := range []bool{true, false} {
		mycelium := newFakeMycelium(t, "", http.StatusNotFound)
		mycelium.SetPage("myc-home", "people/bob.md", "---\ntitle: Bob\ntype: actor\n---\nBob.\n")
		backend := &stepBackend{steps: [][]call{{{id: "d1", name: "memory_delete", input: `{"path":"people/bob.md"}`}}}}
		f := newFixture(t, backend, nil, mycelium.tune)
		memorySpace(t, f, "myc-home")
		runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "forget bob")
		if err != nil {
			t.Fatal(err)
		}
		var card struct {
			Tool string         `json:"tool"`
			Args map[string]any `json:"args"`
		}
		if err := json.Unmarshal(f.waitEvent(t, runID, "INTERRUPT").Payload, &card); err != nil {
			t.Fatal(err)
		}
		if card.Tool != "memory_delete" || card.Args["kind"] != "mcp" || card.Args["server"] != "Mycelium" ||
			!strings.Contains(card.Args["message"].(string), "people/bob.md") {
			t.Fatalf("the card must show Mycelium's question: %+v", card)
		}
		if _, ok := mycelium.Page("myc-home", "people/bob.md"); !ok {
			t.Fatal("the page was deleted before the owner answered")
		}
		if err := f.runner.Decide(context.Background(), 1, runID, "d1", approve); err != nil {
			t.Fatal(err)
		}
		f.waitDone(t, runID)
		if _, kept := mycelium.Page("myc-home", "people/bob.md"); kept == approve {
			t.Fatalf("approve=%t but the page kept=%t", approve, kept)
		}
	}
}
