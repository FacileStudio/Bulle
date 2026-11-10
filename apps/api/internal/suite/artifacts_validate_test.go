package suite

import (
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/encre/encretest"
)

func TestArtifactToolsRefuseWhatEncreRefuses(t *testing.T) {
	log := encretest.Install(t)
	w := newWorld(t)
	tools := teamTools(t, w)

	_, err := runTool(t, tools, "artifact_create", `{"title":"Bad","content":"# Bad\n\nINVALID"}`)
	if err == nil || !strings.Contains(err.Error(), "forbidden character") || !strings.Contains(err.Error(), "fix it") {
		t.Fatalf("create must surface encre's reason, got %v", err)
	}
	if len(w.fakes.arts.bodies) != 0 {
		t.Fatalf("a refused artifact reached Mycelium: %+v", w.fakes.arts.bodies)
	}
	if _, err := runTool(t, tools, "artifact_create", `{"title":"Plan","content":"# Plan\n\nStep one."}`); err != nil {
		t.Fatalf("a valid artifact must go through: %v", err)
	}
	_, err = runTool(t, tools, "artifact_update", `{"artifact_id":"2026-11-06-art-1","content":"INVALID"}`)
	if err == nil || !strings.Contains(err.Error(), "forbidden character") {
		t.Fatalf("update must surface encre's reason, got %v", err)
	}
	if inputs := encretest.Inputs(t, log); len(inputs) != 3 || inputs[1] != "# Plan\n\nStep one." {
		t.Fatalf("encre must see each content once, as sent: %q", inputs)
	}
	if len(w.fakes.arts.bodies) != 1 {
		t.Fatalf("Mycelium must only see the valid write, got %d", len(w.fakes.arts.bodies))
	}
}
