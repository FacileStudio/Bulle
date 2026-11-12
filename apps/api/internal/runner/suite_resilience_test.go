package runner_test

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

func registreDownRun(t *testing.T, tune func(*fakeAgenda), steps ...[]call) calendarRun {
	t.Helper()
	agenda := newFakeAgenda(t)
	tune(agenda)
	backend := &stepBackend{steps: steps}
	shell := newTool(t, "bash", func(string) (string, error) { return "ran", nil })
	f := newFixture(t, backend, []nacelle.Tool{shell}, agenda.tune)
	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "ajoute dans mon agenda lundi soir: soirée test")
	if err != nil {
		t.Fatal(err)
	}
	return calendarRun{f: f, agenda: agenda, backend: backend, runID: runID}
}

// waitSlow is waitDone for a run that spends Bulle's whole series of
// exchange attempts, 7.5 s, before its first turn.
func waitSlow(t *testing.T, run calendarRun) store.Run {
	t.Helper()
	for deadline := time.Now().Add(20 * time.Second); time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
		got, _ := run.f.store.GetRun(context.Background(), run.f.session.UserID, run.runID)
		if got.Status != "" && got.Status != store.RunRunning && !run.f.runner.IsLive(run.runID) {
			return got
		}
	}
	t.Fatal("the run never ended")
	return store.Run{}
}

func TestARunWaitsOutARegistreRedeploy(t *testing.T) {
	run := registreDownRun(t, func(a *fakeAgenda) { a.missing = 2 },
		[]call{{id: "n1", name: "agenda_create_event", input: `{"title":"Soirée test","start":"2026-11-16T19:00:00+01:00"}`}},
	)
	if done := waitSlow(t, run); done.Status != store.RunFinished {
		t.Fatalf("run %s", done.Status)
	}
	if created, _, _ := run.agenda.counts(); len(created) != 1 || created[0] != "Soirée test" {
		t.Fatalf("Registre answering 404 twice must not cost the run its Agenda tools: %v", created)
	}
	if !slices.Contains(run.backend.mounted(), "agenda_create_event") || strings.Contains(run.backend.prompt(), "temporarily unreachable") {
		t.Fatalf("Agenda must be mounted and not reported unreachable: %v", run.backend.mounted())
	}
}

func TestARunToldAnAppIsTemporarilyUnreachable(t *testing.T) {
	run := registreDownRun(t, func(a *fakeAgenda) { a.missing = 1 << 30 },
		[]call{{id: "s1", name: "search_tools", input: `{"query":"calendar"}`}},
	)
	if done := waitSlow(t, run); done.Status != store.RunFinished {
		t.Fatalf("run %s", done.Status)
	}
	if prompt := run.backend.prompt(); !strings.Contains(prompt, "temporarily unreachable: Agenda") || !strings.Contains(prompt, "never say it is not connected") {
		t.Fatalf("the prompt must say Agenda is temporarily unreachable: %s", prompt)
	}
	if found := toolEnds(t, run.f, run.runID)["s1"]["output_preview"].(string); !strings.Contains(found, "Agenda tools are temporarily unavailable") {
		t.Fatalf("search_tools must say Agenda is temporarily unavailable: %q", found)
	}
}

func TestAnAppBackDuringTheRunJoinsIt(t *testing.T) {
	run := registreDownRun(t, func(a *fakeAgenda) { a.missing = 5 },
		[]call{{id: "s1", name: "search_tools", input: `{"query":"event agenda meeting schedule"}`}},
		[]call{{id: "n1", name: "call_tool", input: `{"name":"agenda_create_event","arguments":{"title":"Soirée test","start":"2026-11-16T19:00:00+01:00"}}`}},
	)
	if done := waitSlow(t, run); done.Status != store.RunFinished {
		t.Fatalf("run %s", done.Status)
	}
	ends := toolEnds(t, run.f, run.runID)
	if found := ends["s1"]["output_preview"].(string); !strings.Contains(found, "agenda_create_event") {
		t.Fatalf("search_tools must list Agenda again once Registre answers: %q", found)
	}
	if created, _, _ := run.agenda.counts(); len(created) != 1 || ends["n1"]["ok"] != true {
		t.Fatalf("the event must reach Agenda through call_tool: %v %v", created, ends["n1"])
	}
}

func TestARegistreRefusalStaysFinal(t *testing.T) {
	run := registreDownRun(t, func(a *fakeAgenda) { a.deny = true })
	if done := waitSlow(t, run); done.Status != store.RunFinished {
		t.Fatalf("run %s", done.Status)
	}
	if asked := len(run.agenda.exchanged()); asked != 1 {
		t.Fatalf("access_denied must not be retried: %d exchanges", asked)
	}
	if strings.Contains(run.backend.prompt(), "temporarily unreachable") || slices.Contains(run.backend.mounted(), "search_tools") {
		t.Fatalf("an app the person may not use is not an outage: %v", run.backend.mounted())
	}
}
