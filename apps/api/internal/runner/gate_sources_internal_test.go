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
// targets), so they are deliberately not taint sources.
var ownerReads = []string{
	"agenda_find_free_slots", "agenda_list_calendars", "antenne_list_providers", "antenne_list_targets",
	"antenne_pool_status", "courrier_list_accounts", "echo_list_rooms", "journal_count_logs", "journal_list_alerts",
	"journal_list_apps", "journal_list_saved_queries", "nuage_list_shares", "opus_list_workspaces",
	"sablier_timer_status", "vision_list_sites", "vision_list_spaces",
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
