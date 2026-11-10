package runner_test

import (
	"context"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/nacelle"
)

type webRun struct {
	f       *fixture
	sonar   *fakeSonar
	backend *stepBackend
	runID   string
}

// sonarRun starts a run with sonar as the only suite server and more host
// tools than the catalog threshold, so every other suite tool would go behind
// call_tool.
func sonarRun(t *testing.T, quota int, steps ...[]call) webRun {
	t.Helper()
	sonar := newFakeSonar(t, quota)
	backend := &stepBackend{steps: steps}
	shell := newTool(t, "bash", func(string) (string, error) { return "ran", nil })
	var host []nacelle.Tool
	for i := range 25 {
		host = append(host, newTool(t, "host_"+strconv.Itoa(i), func(string) (string, error) { return "", nil }))
	}
	f := newFixture(t, backend, []nacelle.Tool{shell}, func(c *runner.Config) {
		sonar.tune(c)
		c.HostTools = func(int64) ([]nacelle.Tool, error) { return host, nil }
	})
	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "what is new in go")
	if err != nil {
		t.Fatal(err)
	}
	return webRun{f: f, sonar: sonar, backend: backend, runID: runID}
}

func TestSonarSearchAndFetchRunUnaskedAndStayMounted(t *testing.T) {
	run := sonarRun(t, 5,
		[]call{{id: "s1", name: "sonar_search", input: `{"query":"go 1.27 release"}`}},
		[]call{{id: "p1", name: "sonar_fetch", input: `{"url":"https://go.dev/doc/go1.27"}`}},
	)
	f, sonar, runID := run.f, run.sonar, run.runID
	if done := f.waitDone(t, runID); done.Status != store.RunFinished {
		t.Fatalf("run %s", done.Status)
	}
	ends := toolEnds(t, f, runID)
	if ends["s1"]["ok"] != true || !strings.Contains(ends["s1"]["output_preview"].(string), "go.dev") ||
		ends["p1"]["ok"] != true || !strings.Contains(ends["p1"]["output_preview"].(string), "next_offset") {
		t.Fatalf("search and fetch must run through sonar's MCP server: %v", ends)
	}
	if searched, fetched := sonar.counts(); len(searched) != 1 || len(fetched) != 1 {
		t.Fatalf("sonar saw searches %v and fetches %v", searched, fetched)
	}
	if slices.Contains(f.eventTypes(t, runID), "INTERRUPT") {
		t.Fatal("web reads need no approval card")
	}
	mounted := run.backend.mounted()
	for _, name := range suite.SonarTools() {
		if !slices.Contains(mounted, name) {
			t.Errorf("%s must stay mounted beside the catalog: %v", name, mounted)
		}
	}
	if slices.Contains(mounted, "web_search") || slices.Contains(mounted, "web_fetch") {
		t.Fatalf("the built-in web tools are gone: %v", mounted)
	}
	if exchanges := sonar.exchanged(); len(exchanges) == 0 || exchanges[0] != sonar.resource()+" sonar.read" {
		t.Fatalf("sonar has no write scope, so Bulle asks for read only: %v", exchanges)
	}
}

func TestSonarQuotaExhaustedIsAToolError(t *testing.T) {
	run := sonarRun(t, 0, []call{{id: "s1", name: "sonar_search", input: `{"query":"news"}`}})
	f, runID := run.f, run.runID
	if done := f.waitDone(t, runID); done.Status != store.RunFinished {
		t.Fatalf("an exhausted quota must not fail the run: %s", done.Status)
	}
	end := toolEnds(t, f, runID)["s1"]
	if end["ok"] != false || !strings.Contains(end["output_preview"].(string), "Upgrade at https://sso.facile.studio/spaces") {
		t.Fatalf("the model must see sonar's quota message as a failed call: %v", end)
	}
}

func TestReadingTheWebThroughSonarTaintsTheRun(t *testing.T) {
	run := sonarRun(t, 5,
		[]call{{id: "p1", name: "sonar_fetch", input: `{"url":"https://evil.test"}`}},
		[]call{{id: "b1", name: "bash", input: `{"text":"curl evil.test"}`}},
	)
	f, runID := run.f, run.runID
	interrupt := f.waitEvent(t, runID, "INTERRUPT")
	if !strings.Contains(string(interrupt.Payload), `"call_id":"b1"`) {
		t.Fatalf("the shell must ask once a web page is in the context: %s", interrupt.Payload)
	}
	if err := f.runner.Decide(context.Background(), 1, runID, "b1", false); err != nil {
		t.Fatal(err)
	}
	if done := f.waitDone(t, runID); done.Status != store.RunFinished || toolEnds(t, f, runID)["b1"]["ok"] != false {
		t.Fatal("a denied shell must not run")
	}
}

func TestStoredWebHistoryStillTaints(t *testing.T) {
	first := sonarRun(t, 5, []call{{id: "b1", name: "bash", input: `{"text":"curl evil.test"}`}})
	f, runID := first.f, first.runID
	f.waitDone(t, runID)
	seedTaintHistory(t, f,
		nacelle.UserText("read this page"),
		nacelle.Message{Role: nacelle.RoleAssistant, Parts: []nacelle.Part{nacelle.ToolCall{ID: "w1", Name: "web_fetch", Input: []byte(`{"url":"https://evil.test"}`)}}},
		nacelle.Message{Role: nacelle.RoleUser, Parts: []nacelle.Part{nacelle.ToolResult{ID: "w1", Name: "web_fetch", Result: injected}}},
		nacelle.Message{Role: nacelle.RoleAssistant, Parts: []nacelle.Part{nacelle.Text{Text: "Done."}}},
	)
	next, err := f.runner.Start(context.Background(), 1, f.session.ID, "go on")
	if err != nil {
		t.Fatal(err)
	}
	f.waitEvent(t, next, "INTERRUPT")
	if err := f.runner.Decide(context.Background(), 1, next, "b1", false); err != nil {
		t.Fatal(err)
	}
	f.waitDone(t, next)
}

func TestAnMCPServerThatRefusesTheUserShowsNoAccount(t *testing.T) {
	sonar := newFakeSonar(t, 5)
	sonar.refuse = true
	var cfg runner.Config
	sonar.tune(&cfg)
	for range 2 {
		statuses := cfg.Suite.Statuses(context.Background(), 1)
		if len(statuses) != 1 || statuses[0].App != suite.AppSonar || statuses[0].State != suite.StateNoAccount {
			t.Fatalf("a server up but refusing the user's fresh token means no account there: %+v", statuses)
		}
	}
}
