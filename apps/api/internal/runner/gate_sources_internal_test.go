package runner

import (
	"slices"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
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

var readVerbs = []string{"list", "read", "get", "search", "find", "count", "status", "summary"}

func TestEverySuiteReadIsASourceOrAReviewedOwnerRead(t *testing.T) {
	for app, tools := range suite.MCPToolLists() {
		for _, name := range tools {
			verbs := strings.Split(strings.TrimPrefix(name, app+"_"), "_")
			if !slices.ContainsFunc(verbs, func(verb string) bool { return slices.Contains(readVerbs, verb) }) {
				continue
			}
			if !untrusted(name) && !slices.Contains(ownerReads, name) {
				t.Errorf("%s reads %s data: add it to untrustedSources, or to ownerReads once you checked only the owner writes it", name, app)
			}
		}
	}
}
