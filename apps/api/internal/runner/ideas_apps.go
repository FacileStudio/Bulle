package runner

import (
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

var appToolPrefixes = map[string][]string{
	"mycelium": {"memory_", "artifact_"},
}

var appLabels = map[string]string{
	"courrier": "Courrier", "agenda": "Agenda", "opus": "Opus", "sablier": "Sablier",
	"mycelium": "Mycelium", "nuage": "Nuage", "glouton": "Glouton", store.IdeaAppBulle: "Bulle",
}

var appGuides = map[string]string{
	"courrier": "Look for mail that waits on the user: questions with no reply, threads where the user promised something, " +
		"sent mail with no answer after a few days. Prefer an idea with a ready reply draft.",
	"agenda": "Look at the coming days: meetings to prepare, invitations to answer, clashes, days with no free slot, " +
		"and events that need a follow-up afterwards.",
	"opus": "Look at tasks that are overdue, due soon, blocked or unassigned, and projects that have not moved lately. " +
		"Give one clear next step per idea.",
	"sablier": "Look at the user's tracked time: projects with no time logged lately, a timer left running, " +
		"and work done today that has no time entry.",
	"nuage": "Look at recent files: documents to share, sort, rename or follow up on.",
	"mycelium": "Look at the user's memory pages and artifacts: pages that are stale or contradict what you read, " +
		"and notes worth writing up.",
	"glouton": "Glouton is the user's prospecting tool. Look for these, in this order:\n" +
		"- Lead follow-ups: a contact who replied or opened, a draft never sent, a warm or hot company with no recent touch. Prefer an idea with a ready outreach draft.\n" +
		"- Competitor page changes: a competitor whose site, pricing or offer changed, and what the user could do about it.\n" +
		"- Newly registered competitors: a new company in the user's sector or area worth a look.\n" +
		"- Hunt segments worth adding: a sector, size or region where the best leads cluster and the user does not hunt yet.\n" +
		"Skip a kind your tools show nothing about. Use kind follow_up for lead follow-ups and research for the others.",
}

func appLabel(app string) string {
	if label, ok := appLabels[app]; ok {
		return label
	}
	return app
}

// toolOfApp reports whether a suite tool belongs to app; every tool belongs
// to an unscoped batch.
func toolOfApp(name, app string) bool {
	if app == "" {
		return true
	}
	prefixes, ok := appToolPrefixes[app]
	if !ok {
		prefixes = []string{app + "_"}
	}
	return slices.ContainsFunc(prefixes, func(prefix string) bool { return strings.HasPrefix(name, prefix) })
}

// appGuidance is the prompt section that scopes a batch to one app: what to
// look for there and how to key its fingerprints.
func appGuidance(app string) string {
	if app == "" {
		return ""
	}
	label := appLabel(app)
	guide := "Ideas for " + label + " only. The user reads these cards inside " + label +
		" and runs them there, so propose only what the user does in " + label + ", from " + label + "'s data. " +
		"Start every fingerprint with " + app + ":."
	if extra, ok := appGuides[app]; ok {
		guide += "\n" + extra
	}
	return guide
}

func appOfTool(name string) string {
	for _, app := range store.IdeaApps() {
		if app != store.IdeaAppBulle && toolOfApp(name, app) {
			return app
		}
	}
	return store.IdeaAppBulle
}

func clipSummary(text string) string {
	text = strings.TrimSpace(text)
	if utf8.RuneCountInString(text) <= summaryLimit {
		return text
	}
	return string([]rune(text)[:summaryLimit]) + "..."
}
