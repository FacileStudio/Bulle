package runner

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/nacelle"
)

// ownerReads are the suite reads that return only what the owner set up
// themselves (accounts, calendars, workspaces, their own alerts and
// targets) or the app's own fixed text (a format guide), so they are
// deliberately not taint sources.
var ownerReads = []string{
	"agenda_find_free_slots", "agenda_list_calendars", "antenne_list_providers", "antenne_list_targets",
	"antenne_pool_status", "courrier_list_accounts", "echo_list_rooms", "journal_count_logs", "journal_list_alerts",
	"journal_list_apps", "journal_list_saved_queries", "nuage_list_shares", "opus_list_workspaces",
	"sablier_timer_status", "vision_list_sites", "vision_list_spaces",
	"diapo_format_guide", "clap_format_guide",
}

// thirdPartyVerbs are verbs of tools that are not reads but still hand the
// model other people's text: a summary of their words, a replayed event, a
// target's answer to a check or a test.
var thirdPartyVerbs = []string{"summarize", "replay", "check", "test"}

func TestEverySuiteReadIsASourceOrAReviewedOwnerRead(t *testing.T) {
	for app, tools := range suite.MCPToolLists() {
		for _, name := range tools {
			verbs := strings.Split(strings.TrimPrefix(name, app+"_"), "_")
			if !readName(name) && !slices.ContainsFunc(verbs, func(verb string) bool { return slices.Contains(thirdPartyVerbs, verb) }) {
				continue
			}
			if !untrusted(name) && !slices.Contains(ownerReads, name) {
				t.Errorf("%s reads %s data: add it to untrustedSources, or to ownerReads once you checked only the owner writes it", name, app)
			}
		}
	}
}

func TestEverySuiteWriteAsksInATaintedRunUnlessItsServerConfirms(t *testing.T) {
	g := trustGate(nacelle.UserText("read my mail"))
	g.observe(context.Background(), nacelle.HookEvent{Point: nacelle.AfterToolCall, Call: nacelle.Call{Name: "courrier_read_thread"}})
	writes := 0
	for _, tools := range suite.MCPToolLists() {
		for _, name := range tools {
			if name != sonarFetchTool && !readName(name) {
				writes++
			}
			checkTaintedAsk(t, g, name)
		}
	}
	if writes < 60 {
		t.Fatalf("only %d suite writes checked", writes)
	}
	for _, name := range serverConfirmed {
		if !listedTools[name] || readName(name) {
			t.Errorf("%s is server confirmed but is not a listed suite write", name)
		}
	}
}

// checkTaintedAsk fails when a listed suite tool asks in a tainted run and
// should not, or does not and should.
func checkTaintedAsk(t *testing.T, g *gate, name string) {
	t.Helper()
	if name == sonarFetchTool {
		return
	}
	asks := g.asks(name, json.RawMessage(`{}`))
	switch {
	case readName(name) && asks:
		t.Errorf("%s reads, but asks in a tainted run", name)
	case !readName(name) && asks == serverConfirms(name):
		t.Errorf("%s writes: it must ask in a tainted run unless its server confirms it (serverConfirmed %t)", name, serverConfirms(name))
	}
}

func TestAMountedWriteBulleDoesNotListAsksInATaintedRun(t *testing.T) {
	g := trustGate(nacelle.UserText("plan the sprint"))
	tools := []nacelle.Tool{
		specTool{nacelle.Spec{Name: "opus_archive_board"}},
		specTool{nacelle.Spec{Name: "opus_peek_board", ReadOnly: true}},
	}
	g.addWrites(mountedWrites(tools))
	if g.asks("opus_archive_board", nil) {
		t.Fatal("a clean run asked for a suite write")
	}
	g.observe(context.Background(), nacelle.HookEvent{Point: nacelle.AfterToolCall, Call: nacelle.Call{Name: "opus_read_task"}})
	if !g.asks("opus_archive_board", nil) || g.asks("opus_peek_board", nil) {
		t.Fatal("a tainted run must ask for an unlisted mounted write and leave a read-only tool unasked")
	}
}

func TestAMountedReadBulleDoesNotListIsASource(t *testing.T) {
	tools := []nacelle.Tool{
		specTool{nacelle.Spec{Name: "opus_peek_board", ReadOnly: true}},
		specTool{nacelle.Spec{Name: "opus_read_task", ReadOnly: true}},
		specTool{nacelle.Spec{Name: "opus_archive_board"}},
	}
	if got := unlistedReads(tools); !slices.Equal(got, []string{"opus_peek_board"}) {
		t.Fatalf("unlisted reads = %v, want only the unlisted read-only tool", got)
	}
	g := trustGate(nacelle.UserText("plan the sprint"))
	g.addSources(mountedSources(tools))
	if g.asks(shellTool, nil) {
		t.Fatal("a clean run asked for the shell")
	}
	g.observe(context.Background(), nacelle.HookEvent{Point: nacelle.AfterToolCall, Call: nacelle.Call{Name: "opus_peek_board"}})
	if !g.asks(shellTool, nil) {
		t.Fatal("a run that read an unlisted tool must be tainted")
	}
}

func TestTheAppContractToolsAreClassifiedAsIntended(t *testing.T) {
	g := trustGate(nacelle.UserText("plan my week"))
	if g.asks("agenda_propose_plan", nil) {
		t.Fatal("a clean run asked for the plan draft")
	}
	for _, name := range []string{"agenda_list_plans", "vision_list_anomalies", "vision_compare_periods"} {
		if !untrusted(name) || !readName(name) || g.asks(name, nil) {
			t.Errorf("%s must be a read that is also an untrusted source", name)
		}
	}
	g.observe(context.Background(), nacelle.HookEvent{Point: nacelle.AfterToolCall, Call: nacelle.Call{Name: "agenda_list_plans"}})
	if !g.asks("agenda_propose_plan", nil) || serverConfirms("agenda_propose_plan") || readName("agenda_propose_plan") {
		t.Fatal("a run that read a plan must ask before it saves another")
	}
}

// sceneReads, sceneWrites and sceneConfirmed are every Diapo and Clap tool,
// by what the gate must make of it.
var (
	sceneReads = []string{"diapo_list_decks", "diapo_get_deck", "diapo_format_guide", "clap_list_projects", "clap_get_project",
		"clap_format_guide", "clap_list_assets", "clap_render_status"}
	sceneWrites = []string{"diapo_create_deck", "diapo_write_deck", "diapo_edit_deck", "diapo_export",
		"clap_create_project", "clap_write_project", "clap_edit_project", "clap_timeline_edit", "clap_render"}
	sceneConfirmed = []string{"diapo_delete_deck", "clap_delete_project"}
)

func TestDiapoAndClapToolsAreClassifiedAsIntended(t *testing.T) {
	every := slices.Concat(sceneReads, sceneWrites, sceneConfirmed)
	if listed := len(suite.DiapoTools()) + len(suite.ClapTools()); len(every) != listed {
		t.Fatalf("%d Diapo and Clap tools are listed, %d classified here", listed, len(every))
	}
	clean := trustGate(nacelle.UserText("make me a deck"))
	for _, name := range every {
		if !listedTools[name] || clean.asks(name, nil) {
			t.Errorf("%s must be listed and run unasked in a clean run", name)
		}
	}
	for _, name := range sceneReads {
		if !readName(name) || untrusted(name) == strings.HasSuffix(name, "_format_guide") {
			t.Errorf("%s must read, and taint unless it is the format guide", name)
		}
	}
	clean.observe(context.Background(), nacelle.HookEvent{Point: nacelle.AfterToolCall, Call: nacelle.Call{Name: "diapo_format_guide"}})
	if clean.asks("diapo_create_deck", nil) {
		t.Fatal("the format guide is the app's own text and must not taint the run")
	}
}

func TestReadingADeckOrAProjectTaintsTheRun(t *testing.T) {
	for _, source := range []string{"diapo_get_deck", "clap_get_project"} {
		g := trustGate(nacelle.UserText("fix the title"))
		g.observe(context.Background(), nacelle.HookEvent{Point: nacelle.AfterToolCall, Call: nacelle.Call{Name: source}})
		for _, name := range sceneWrites {
			if !g.asks(name, nil) {
				t.Errorf("%s must ask once %s tainted the run", name, source)
			}
		}
		for _, name := range slices.Concat(sceneReads, sceneConfirmed) {
			if g.asks(name, nil) {
				t.Errorf("%s must not raise Bulle's card after %s: it reads, or its app confirms", name, source)
			}
		}
	}
}
