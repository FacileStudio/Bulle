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
	"courrier": "Courrier is the user's mail. Look for these, in this order:\n" +
		"- Replies owed: a question or request to the user with no reply yet. Prefer an idea with a ready reply draft.\n" +
		"- Promises: a thread where the user said they would send or do something and has not.\n" +
		"- Silence after sending: mail the user sent a few days ago that got no answer; draft a short nudge.\n" +
		"- Inbox upkeep: a sender or newsletter that keeps piling up unread and is worth archiving.\n" +
		"Skip a kind your tools show nothing about. Use kind mail for replies and nudges, follow_up for promises and other for upkeep.",
	"agenda": "Agenda is the user's calendar. Look at the next seven days, in this order:\n" +
		"- Meetings to prepare: an event in the next two days with other people or a topic; say what to read or bring.\n" +
		"- Conflicts: overlapping events, or back-to-back events in different places; say which one to move and where.\n" +
		"- Invitations still waiting for the user's answer.\n" +
		"- Follow-ups: a meeting that just ended and needs a recap or a next date.\n" +
		"- Full days: a day with no free slot (agenda_find_free_slots shows them); suggest a focus block on a lighter day.\n" +
		"Skip a kind your tools show nothing about. Use kind event, and follow_up for recaps.",
	"opus": "Opus is the user's task manager. Look for these, in this order:\n" +
		"- Overdue tasks: give the one next step, or a realistic new due date.\n" +
		"- Tasks due in the next three days that have not started.\n" +
		"- Unassigned tasks in a shared project: say who could take each one.\n" +
		"- Blocked tasks, and projects with no change in two weeks: name what unblocks them.\n" +
		"Give one clear next step per idea. Skip a kind your tools show nothing about. Use kind task, and follow_up for stalled projects.",
	"sablier": "Sablier is the user's time tracker. Look for these, in this order:\n" +
		"- Timers left running: a timer running for hours, or since an earlier day; propose stopping it.\n" +
		"- A timer on the wrong project, judging by the project names and the time of day.\n" +
		"- Untracked time: projects the user tracks often that have no time logged lately; propose starting a timer.\n" +
		"Skip a kind your tools show nothing about. Use kind task.",
	"nuage": "Nuage is the user's file storage. Look at recent files, in this order:\n" +
		"- Files to organize: loose files at the top level, duplicates, or names that say nothing; name the folder or the new name.\n" +
		"- Files to share: a recent quote, contract or report that looks meant for someone; name who and why.\n" +
		"- Documents to follow up: a draft left unfinished, or a document that asks for an answer.\n" +
		"Your tools read and save files but cannot move or share them, so a run saves what helps (an index, a summary to send) and the card says what the user should move or share.\n" +
		"Skip a kind your tools show nothing about. Use kind other, and follow_up for documents to follow up.",
	"mycelium": "Mycelium is the user's notes and memory. Look for these, in this order:\n" +
		"- Notes to update: a page that is stale or contradicts what you read elsewhere; say what changed.\n" +
		"- Notes to merge: two pages on the same topic.\n" +
		"- Notes worth writing: an artifact or a recurring topic with no page yet.\n" +
		"Skip a kind your tools show nothing about. Use kind research for updates and other for the rest.",
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
