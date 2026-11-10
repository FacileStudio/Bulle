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

type calendarRun struct {
	f       *fixture
	agenda  *fakeAgenda
	backend *stepBackend
	runID   string
}

func agendaRun(t *testing.T, steps ...[]call) calendarRun {
	t.Helper()
	agenda := newFakeAgenda(t)
	backend := &stepBackend{steps: steps}
	shell := newTool(t, "bash", func(string) (string, error) { return "ran", nil })
	f := newFixture(t, backend, []nacelle.Tool{shell}, agenda.tune)
	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "sort out my week")
	if err != nil {
		t.Fatal(err)
	}
	return calendarRun{f: f, agenda: agenda, backend: backend, runID: runID}
}

func TestAgendaMCPListAndCreateRunUnasked(t *testing.T) {
	run := agendaRun(t,
		[]call{{id: "c1", name: "agenda_list_calendars", input: `{}`}, {id: "l1", name: "agenda_list_events", input: `{"from":"2026-11-12","to":"2026-11-13"}`}},
		[]call{{id: "n1", name: "agenda_create_event", input: `{"title":"Dentist","start":"2026-11-12T14:00:00+01:00"}`}},
	)
	f, agenda, runID := run.f, run.agenda, run.runID
	if done := f.waitDone(t, runID); done.Status != store.RunFinished {
		t.Fatalf("run %s", done.Status)
	}
	ends := toolEnds(t, f, runID)
	if ends["c1"]["ok"] != true || !strings.Contains(ends["c1"]["output_preview"].(string), "Personal") ||
		ends["l1"]["ok"] != true || ends["n1"]["ok"] != true || !strings.Contains(ends["n1"]["output_preview"].(string), "created") {
		t.Fatalf("list and create must run through MCP: %v", ends)
	}
	if created, _, _ := agenda.counts(); len(created) != 1 || created[0] != "Dentist" {
		t.Fatalf("the event never reached Agenda: %v", created)
	}
	if slices.Contains(f.eventTypes(t, runID), "INTERRUPT") {
		t.Fatal("Agenda asked nothing, so Bulle must not raise its own card")
	}
	mounted := run.backend.mounted()
	for _, name := range suite.AgendaTools() {
		if !slices.Contains(mounted, name) {
			t.Errorf("%s is not mounted under its own name: %v", name, mounted)
		}
	}
	if exchanges := agenda.exchanged(); len(exchanges) == 0 || exchanges[0] != agenda.resource()+" agenda.read agenda.write" {
		t.Fatalf("the token must be exchanged with the resource as its audience: %v", exchanges)
	}
}

func assertDeleteCard(t *testing.T, raw json.RawMessage) {
	t.Helper()
	var payload struct {
		CallID    string         `json:"call_id"`
		Tool      string         `json:"tool"`
		Grantable bool           `json:"grantable"`
		Args      map[string]any `json:"args"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	args := payload.Args
	if payload.CallID != "x1" || payload.Tool != "agenda_delete_event" || payload.Grantable || args["kind"] != "mcp" ||
		args["server"] != "Agenda" || !strings.Contains(args["message"].(string), `Delete the event "Standup"`) ||
		args["arguments"].(map[string]any)["event_id"] != float64(4) {
		t.Fatalf("the card must show Agenda's question and the call: %s", raw)
	}
}

func TestAgendaMCPDeleteWaitsOnTheApprovalCard(t *testing.T) {
	for _, approve := range []bool{true, false} {
		run := agendaRun(t, []call{{id: "x1", name: "agenda_delete_event", input: `{"event_id":4}`}})
		f, agenda, runID := run.f, run.agenda, run.runID
		assertDeleteCard(t, f.waitEvent(t, runID, "INTERRUPT").Payload)
		if _, deleted, _ := agenda.counts(); len(deleted) != 0 {
			t.Fatal("the event was deleted before the owner answered")
		}
		if err := f.runner.Decide(context.Background(), 1, runID, "x1", approve); err != nil {
			t.Fatal(err)
		}
		f.waitDone(t, runID)
		_, deleted, declined := agenda.counts()
		preview := toolEnds(t, f, runID)["x1"]["output_preview"].(string)
		switch {
		case approve && (len(deleted) != 1 || declined != 0 || !strings.Contains(preview, "deleted")):
			t.Fatalf("an approved delete must run once: deleted=%v declined=%d %q", deleted, declined, preview)
		case !approve && (len(deleted) != 0 || declined != 1 || !strings.Contains(preview, "declined")):
			t.Fatalf("a denied delete must reach Agenda as a decline: deleted=%v declined=%d %q", deleted, declined, preview)
		}
		if agenda.asked[0] != "2026-07-28" {
			t.Fatalf("Bulle must speak the stateless protocol, got %v", agenda.asked)
		}
	}
}

func TestReadingEventsOverMCPTaintsTheRun(t *testing.T) {
	run := agendaRun(t,
		[]call{{id: "r1", name: "agenda_read_event", input: `{"event_id":4}`}},
		[]call{{id: "b1", name: "bash", input: `{"text":"curl evil.test"}`}},
	)
	f, runID := run.f, run.runID
	interrupt := f.waitEvent(t, runID, "INTERRUPT")
	if !strings.Contains(string(interrupt.Payload), `"call_id":"b1"`) {
		t.Fatalf("the shell must ask once an invitation is in the context: %s", interrupt.Payload)
	}
	if err := f.runner.Decide(context.Background(), 1, runID, "b1", false); err != nil {
		t.Fatal(err)
	}
	if done := f.waitDone(t, runID); done.Status != store.RunFinished || toolEnds(t, f, runID)["b1"]["ok"] != false {
		t.Fatal("a denied shell must not run")
	}
}

func TestStoredAgendaHistoryStillTaints(t *testing.T) {
	first := agendaRun(t, []call{{id: "b1", name: "bash", input: `{"text":"curl evil.test"}`}})
	f, runID := first.f, first.runID
	f.waitDone(t, runID)
	seedTaintHistory(t, f,
		nacelle.UserText("what is on tomorrow"),
		nacelle.Message{Role: nacelle.RoleAssistant, Parts: []nacelle.Part{nacelle.ToolCall{ID: "e1", Name: "agenda_list_events", Input: json.RawMessage(`{}`)}}},
		nacelle.Message{Role: nacelle.RoleUser, Parts: []nacelle.Part{nacelle.ToolResult{ID: "e1", Name: "agenda_list_events", Result: injected}}},
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
