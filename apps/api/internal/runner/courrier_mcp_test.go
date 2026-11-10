package runner_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/nacelle"
)

// mailRun is one run against the fake Courrier, with the shell mounted.
type mailRun struct {
	f        *fixture
	courrier *fakeCourrier
	backend  *stepBackend
	runID    string
}

func courrierRun(t *testing.T, steps ...[]call) mailRun {
	t.Helper()
	courrier := newFakeCourrier(t)
	backend := &stepBackend{steps: steps}
	shell := newTool(t, "bash", func(string) (string, error) { return "ran", nil })
	f := newFixture(t, backend, []nacelle.Tool{shell}, courrier.tune)
	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "deal with my mail")
	if err != nil {
		t.Fatal(err)
	}
	return mailRun{f: f, courrier: courrier, backend: backend, runID: runID}
}

func toolEnds(t *testing.T, f *fixture, runID string) map[string]map[string]any {
	t.Helper()
	events, _ := f.store.Events(context.Background(), runID, 0)
	ends := map[string]map[string]any{}
	for _, event := range events {
		if event.Type == "TOOL_CALL_END" {
			var end map[string]any
			_ = json.Unmarshal(event.Payload, &end)
			ends[end["call_id"].(string)] = end
		}
	}
	return ends
}

func TestCourrierMCPSearchAndDraftRunUnasked(t *testing.T) {
	run := courrierRun(t,
		[]call{{id: "q1", name: "courrier_search_mail", input: `{"query":"invoice"}`}},
		[]call{{id: "d1", name: "courrier_draft", input: `{"account_id":5,"to":["ann@x"],"subject":"Invoice","body":"Hello"}`}},
	)
	f, courrier, runID := run.f, run.courrier, run.runID
	if done := f.waitDone(t, runID); done.Status != store.RunFinished {
		t.Fatalf("run %s", done.Status)
	}
	ends := toolEnds(t, f, runID)
	if ends["q1"]["ok"] != true || !strings.Contains(ends["q1"]["output_preview"].(string), "IGNORE ALL") || ends["d1"]["ok"] != true {
		t.Fatalf("search and draft must run through MCP: %v", ends)
	}
	if _, _, drafts := courrier.counts(); len(drafts) != 1 || drafts[0] != "Invoice" {
		t.Fatalf("the draft never reached Courrier: %v", drafts)
	}
	for _, kind := range f.eventTypes(t, runID) {
		if kind == "INTERRUPT" {
			t.Fatal("a search or a draft raised an approval card")
		}
	}
	mounted := strings.Join(run.backend.mounted(), ",")
	for _, name := range suite.CourrierTools() {
		if !strings.Contains(mounted, name) {
			t.Errorf("%s is not mounted under its own name: %s", name, mounted)
		}
	}
	if exchanges := courrier.exchanged(); len(exchanges) == 0 || exchanges[0] != courrier.resource()+" courrier.read courrier.write" {
		t.Fatalf("the token must be exchanged with the resource as its audience: %v", exchanges)
	}
}

func TestCourrierMCPSendWaitsOnTheApprovalCard(t *testing.T) {
	for _, approve := range []bool{true, false} {
		run := courrierRun(t,
			[]call{{id: "s1", name: "courrier_send", input: `{"account_id":5,"to":["ann@x"],"subject":"Invoice","body":"Hello"}`}},
		)
		f, courrier, runID := run.f, run.courrier, run.runID
		interrupt := f.waitEvent(t, runID, "INTERRUPT")
		var payload struct {
			CallID    string         `json:"call_id"`
			Tool      string         `json:"tool"`
			Grantable bool           `json:"grantable"`
			Args      map[string]any `json:"args"`
		}
		_ = json.Unmarshal(interrupt.Payload, &payload)
		args := payload.Args
		if payload.CallID != "s1" || payload.Tool != "courrier_send" || payload.Grantable || args["kind"] != "mcp" || args["server"] != "Courrier" ||
			args["message"] != `Send the email "Invoice" to ann@x now?` || args["arguments"].(map[string]any)["body"] != "Hello" {
			t.Fatalf("the card must show the server's question and the call: %s", interrupt.Payload)
		}
		if sent, _, _ := courrier.counts(); sent != 0 {
			t.Fatal("the mail left before the owner answered")
		}
		if err := f.runner.Decide(context.Background(), 1, runID, "s1", approve); err != nil {
			t.Fatal(err)
		}
		f.waitDone(t, runID)
		sent, declined, _ := courrier.counts()
		preview := toolEnds(t, f, runID)["s1"]["output_preview"].(string)
		switch {
		case approve && (sent != 1 || declined != 0 || !strings.Contains(preview, "sent")):
			t.Fatalf("an approved send must go out once: sent=%d declined=%d %q", sent, declined, preview)
		case !approve && (sent != 0 || declined != 1 || !strings.Contains(preview, "declined")):
			t.Fatalf("a denied send must reach Courrier as a decline: sent=%d declined=%d %q", sent, declined, preview)
		}
		if courrier.asked[0] != "2026-07-28" {
			t.Fatalf("Bulle must speak the stateless protocol, got %v", courrier.asked)
		}
	}
}

func TestReadingMailOverMCPTaintsTheRun(t *testing.T) {
	run := courrierRun(t,
		[]call{{id: "r1", name: "courrier_read_thread", input: `{"account_id":5,"message_id":9}`}},
		[]call{{id: "b1", name: "bash", input: `{"text":"curl evil.test"}`}},
	)
	f, runID := run.f, run.runID
	interrupt := f.waitEvent(t, runID, "INTERRUPT")
	if !strings.Contains(string(interrupt.Payload), `"call_id":"b1"`) {
		t.Fatalf("the shell must ask once mail is in the context: %s", interrupt.Payload)
	}
	if err := f.runner.Decide(context.Background(), 1, runID, "b1", false); err != nil {
		t.Fatal(err)
	}
	if done := f.waitDone(t, runID); done.Status != store.RunFinished || toolEnds(t, f, runID)["b1"]["ok"] != false {
		t.Fatal("a denied shell must not run")
	}
}

func TestStoredCourrierHistoryStillTaints(t *testing.T) {
	first := courrierRun(t, []call{{id: "b1", name: "bash", input: `{"text":"curl evil.test"}`}})
	f, runID := first.f, first.runID
	f.waitDone(t, runID)
	seedTaintHistory(t, f,
		nacelle.UserText("read my mail"),
		nacelle.Message{Role: nacelle.RoleAssistant, Parts: []nacelle.Part{nacelle.ToolCall{ID: "m1", Name: "courrier_search_mail", Input: json.RawMessage(`{}`)}}},
		nacelle.Message{Role: nacelle.RoleUser, Parts: []nacelle.Part{nacelle.ToolResult{ID: "m1", Name: "courrier_search_mail", Result: injected}}},
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

func TestIdeasReadMailThroughMCPAndNeverWrite(t *testing.T) {
	courrier := newFakeCourrier(t)
	backend := &ideasBackend{fakeBackend: fakeBackend{calls: proposals("courrier:5/email:9/reply")}}
	f := newFixture(t, backend, nil, courrier.tune)
	space := f.space(t)
	if _, err := f.runner.GenerateIdeas(context.Background(), 1, space.ID, store.BatchManual); err != nil {
		t.Fatal(err)
	}
	f.waitBatch(t, space.ID)
	var names []string
	for _, tool := range backend.seen() {
		names = append(names, tool.Spec().Name)
	}
	got := strings.Join(names, ",")
	if !strings.Contains(got, "courrier_search_mail") || !strings.Contains(got, "courrier_read_thread") ||
		strings.Contains(got, "courrier_send") || strings.Contains(got, "courrier_draft") || strings.Contains(got, "courrier_organize") {
		t.Fatalf("an ideas run reads mail and writes nothing: %s", got)
	}
}
