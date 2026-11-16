package runner

import (
	"context"
	"encoding/json"
	"slices"
	"strings"

	"github.com/FacileStudio/Bulle/core/compaction"
	"github.com/FacileStudio/nacelle"
)

const (
	browserNavigateTool = "browser_navigate"
	browserClickTool    = "browser_click"
	browserTypeTool     = "browser_type"
	sonarFetchTool      = "sonar_fetch"
	memoryWriteTool     = "memory_write"
)

// untrustedSources is every tool whose result carries text that somebody
// other than the owner wrote: mail, the web, the browser, shared files, tasks
// and time, prospects imported or audited from the web, the other suite
// apps' shared data, meeting summaries written from other people's words,
// and an alert target's or a replayed event's answer. Memory is not a source:
// a run that read any of these, or the shell or a file, asks before every
// memory write, and the memory pass learns only from the owner's own messages
// once a chat is tainted, so no third party's text reaches a memory page
// unseen by the owner. A name ending in "_" covers every tool
// with that prefix. Old and new names both stay listed, so a rename never
// quietly drops a source. An MCP tool that is read-only and open world is a
// source too (see openWorldSources), and the suite servers' are listed by
// name so a stored transcript taints before its server is reached.
var untrustedSources = []string{
	"courrier_inbox", "courrier_read", "courrier_search", "courrier_list_inbox", "courrier_get_email",
	"courrier_search_mail", "courrier_read_thread",
	"agenda_list_events", "agenda_read_event", "agenda_list_plans",
	"web_fetch", "web_search", "sonar_search", "sonar_fetch", "browser_",
	"files_list", "files_read", "files_search", "nuage_list", "nuage_read", "nuage_search",
	"nuage_list_folder", "nuage_read_file",
	"opus_list_projects", "opus_search_tasks", "opus_read_task",
	"sablier_list_entries", "sablier_list_projects", "sablier_weekly_summary",
	"glouton_contacts_search", "glouton_company_get", "glouton_person_get", "glouton_outreach_list", "glouton_suggestions_list",
	"plume_list_documents", "plume_get_document", "plume_list_clients", "capsule_get_status",
	"journal_search_logs", "journal_read_log", "antenne_list_events", "antenne_get_event", "antenne_test_provider",
	"antenne_check_target", "antenne_test_target", "antenne_replay_event",
	"vision_get_stats", "vision_list_goals", "vision_list_anomalies", "vision_compare_periods", "echo_list_meetings", "echo_get_meeting", "echo_read_transcript",
	"echo_summarize_meeting",
	"toile_list_boards", "toile_read_board", "toile_list_comments",
	"docs_search", "docs_list_tree", "docs_read", "docs_read_blocks", "docs_list_trash",
	"feuilles_list_workbooks", "feuilles_get_workbook", "feuilles_read_range",
	"diapo_list_decks", "diapo_get_deck",
	"clap_list_projects", "clap_get_project", "clap_list_assets", "clap_render_status",
}

// softSources read the computer: a command's output, a file's text and the
// names of files may come from a cloned repository or a downloaded page.
// They taint softly, so the run's links, clicks, typing and memory writes
// follow the tainted rules while the shell and the suite stay unasked, or a
// coding run would raise a card on every command. The shell is unasked only
// until the run reads private data on a computer with open egress (see
// privateReads).
var softSources = []string{shellTool, "read_file", "grep", "glob", "list_directory"}

// artifactWrites save into the space's Mycelium, where the other members of
// a shared space read them, so a tainted run asks before each.
var artifactWrites = []string{"artifact_create", "artifact_update"}

// shellTool runs commands in the computer.
const shellTool = "bash"

// linkBudget is how many links a tainted run follows unasked when it saw them
// only in a tool's result, browser clicks included. Each choice among the
// links a page lists can carry a few bits of what the run read, so the count
// is bounded, across every run of the session (see seed).
const linkBudget = 3

// linkScope keys the owner's grant to follow more such links in a run, for
// sonar_fetch, browser_navigate and browser_click alike.
const linkScope = "links from results"

type taint int

const (
	clean taint = iota
	soft
	tainted
)

type ask int

const (
	askNo ask = iota
	askCard
	askLinks
	askTaint
)

// trust is what the gate knows about where the run's context came from: how
// far untrusted text reached it, the URLs the owner wrote and the ones a
// source returned, how many of those the run followed unasked, and the
// mounted tools that read or write beyond Bulle.
type trust struct {
	level    taint
	marked   bool
	own      seenURLs
	seen     seenURLs
	followed int
	sources  map[string]bool
	writes   map[string]bool
	wrote    bool
	private  bool
	network  string
}

func untrusted(name string) bool {
	return slices.ContainsFunc(untrustedSources, func(source string) bool {
		if strings.HasSuffix(source, "_") {
			return strings.HasPrefix(name, source)
		}
		return name == source
	})
}

// taintOf is how far a result of name taints the run. The caller holds mu.
func (g *gate) taintOf(name string) taint {
	switch {
	case g.sourceLocked(name):
		return tainted
	case slices.Contains(softSources, name):
		return soft
	}
	return clean
}

// seed reads the conversation the run starts from. A call or a result of an
// untrusted source, or an attachment, anywhere in it taints the run, since
// that text is still in the model's context; a stored shell or file result
// taints it softly. A result carries the name of the tool that ran, which is
// how a source reached through call_tool is caught. The URLs in the owner's
// own text and in the sources' results are kept for the fetch and navigate
// check, and the links earlier runs followed from results once the chat was
// tainted count against this run's budget, so a new owner message does not
// refill it. A call run through call_tool counts as the tool it routed to.
// marked is the session's stored taint flag, set by a soft run too. When it
// is set but the history reads clean, or reads soft behind a compaction
// ledger, the fold may have taken a source's turns away, so the run starts
// tainted with the link budget spent. An unfolded soft chat stays soft. A
// message that carries the idea note (see ideaStarted) was written by an
// agent from third-party text: it taints the run, and its URLs are not the
// owner's.
func (g *gate) seed(history []nacelle.Message, marked bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	own, seen := seenURLs{}, seenURLs{}
	level, followed := clean, 0
	for _, message := range history {
		var spent int
		level, spent = g.readMessage(own, seen, level, message)
		followed += spent
	}
	if marked && (level == clean || level == soft && slices.ContainsFunc(history, compaction.IsLedger)) {
		level, followed = tainted, linkBudget
	}
	g.trust = trust{level: level, marked: marked, own: own, seen: seen, followed: followed, sources: g.trust.sources, writes: g.trust.writes, wrote: g.trust.wrote, private: g.trust.private, network: g.trust.network}
}

// hooks taint the run as soon as an untrusted source returns, before the
// model reads what it returned.
func (g *gate) hooks() map[nacelle.HookPoint][]nacelle.Hook {
	return map[nacelle.HookPoint][]nacelle.Hook{nacelle.AfterToolCall: {g.observe}}
}

func (g *gate) observe(ctx context.Context, ev nacelle.HookEvent) nacelle.HookResult {
	g.mu.Lock()
	level := g.taintOf(ev.Call.Name)
	if slices.Contains(privateReads, ev.Call.Name) {
		g.trust.private = true
	}
	if level != clean {
		g.trust.level = max(g.trust.level, level)
		if g.trust.seen == nil {
			g.trust.seen = seenURLs{}
		}
		g.trust.seen.addResult(ev.Result)
	}
	g.mu.Unlock()
	g.markTainted(ctx)
	return nacelle.HookResult{}
}

// asks reports whether a call would wait for the owner now, without
// counting it against the run's links.
func (g *gate) asks(name string, input json.RawMessage) bool {
	return g.need(name, input, false) != askNo
}

// need is whether a call waits for the owner: an outward tool always. Once
// third-party text reached the run, even through the shell or a file, a
// browser_type, a memory write, a browser_click past the run's linkBudget,
// and a browser_navigate or a sonar_fetch whose URL the owner did not write
// and that was never seen verbatim in a source's result or is past the
// budget. A job run fetches unasked only while no source tainted it. Once a source tainted the run, also a shell command,
// an artifact write outside a job run, and every suite write its server does
// not confirm itself. With spend set, a link or click followed unasked counts
// against the budget. A memory or artifact delete is Mycelium's to confirm.
// A soft-tainted run on an open computer also asks before the shell once it
// read memory or past chats. A card the taint raises is askTaint, which taintScope may let the owner
// grant for the run.
func (g *gate) need(name string, input json.RawMessage, spend bool) ask {
	if g.outward[name] {
		return askCard
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.trust.level == clean {
		return askNo
	}
	switch {
	case name == sonarFetchTool && g.grants[jobScope] && g.trust.level != tainted:
		return askNo
	case name == browserNavigateTool || name == sonarFetchTool:
		return g.linkLocked(input, spend)
	case name == browserClickTool:
		return g.budgetLocked(spend)
	case name == browserTypeTool || name == memoryWriteTool:
		return askTaint
	}
	if g.trust.level != tainted {
		return g.softNeedLocked(name)
	}
	if name == shellTool || (slices.Contains(artifactWrites, name) && !g.grants[jobScope]) || (g.writeLocked(name) && !serverConfirms(name)) {
		return askTaint
	}
	return askNo
}
