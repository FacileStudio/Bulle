package runner

import (
	"context"
	"encoding/json"
	"slices"
	"strings"

	"github.com/FacileStudio/nacelle"
)

const (
	browserNavigateTool = "browser_navigate"
	sonarFetchTool      = "sonar_fetch"
	memoryWriteTool     = "memory_write"
)

// untrustedSources is every tool whose result carries text that somebody
// other than the owner wrote: mail, the web, the browser, shared files, tasks
// and time, prospects imported or audited from the web, the other suite
// apps' shared data. Memory is not a source: a tainted run asks before every
// memory write, and the memory pass learns only from the owner's own messages
// once a chat is tainted, so no third party's text reaches a memory page. A
// name ending in "_" covers every tool with that prefix. Old and new names
// both stay listed, so a rename never quietly drops a source. An MCP tool
// that is read-only and open world is a source too (see openWorldSources),
// and the suite servers' are listed by name so a stored transcript taints
// before its server is reached.
var untrustedSources = []string{
	"courrier_inbox", "courrier_read", "courrier_search", "courrier_list_inbox", "courrier_get_email",
	"courrier_search_mail", "courrier_read_thread",
	"agenda_list_events", "agenda_read_event",
	"web_fetch", "web_search", "sonar_search", "sonar_fetch", "browser_",
	"files_list", "files_read", "files_search", "nuage_list", "nuage_read", "nuage_search",
	"nuage_list_folder", "nuage_read_file",
	"opus_list_projects", "opus_search_tasks", "opus_read_task",
	"sablier_list_entries", "sablier_list_projects", "sablier_weekly_summary",
	"glouton_contacts_search", "glouton_company_get", "glouton_person_get", "glouton_outreach_list", "glouton_suggestions_list",
	"plume_list_documents", "plume_get_document", "plume_list_clients", "capsule_get_status",
	"journal_search_logs", "journal_read_log", "antenne_list_events", "antenne_get_event", "antenne_test_provider",
	"vision_get_stats", "vision_list_goals", "echo_list_meetings", "echo_get_meeting", "echo_read_transcript",
	"toile_list_boards", "toile_read_board", "toile_list_comments",
	"docs_search", "docs_list_tree", "docs_read", "docs_read_blocks", "docs_list_trash",
	"feuilles_list_workbooks", "feuilles_get_workbook", "feuilles_read_range",
}

// shellTool runs commands in the computer, which has open egress.
const shellTool = "bash"

// trust is what the gate knows about where the run's context came from:
// whether untrusted text reached it, and the URLs the owner wrote or an
// untrusted source returned.
type trust struct {
	tainted bool
	seen    seenURLs
	sources map[string]bool
}

func untrusted(name string) bool {
	return slices.ContainsFunc(untrustedSources, func(source string) bool {
		if strings.HasSuffix(source, "_") {
			return strings.HasPrefix(name, source)
		}
		return name == source
	})
}

// seed reads the conversation the run starts from. A call or a result of an
// untrusted source, or an attachment, anywhere in it taints the run, since
// that text is still in the model's context. A result carries the name of the
// tool that ran, which is how a source reached through call_tool is caught.
// The URLs in the owner's own text and in the sources' results are kept for
// the fetch and navigate check.
func (g *gate) seed(history []nacelle.Message) {
	g.mu.Lock()
	defer g.mu.Unlock()
	seen := seenURLs{}
	tainted := false
	for _, message := range history {
		for _, part := range message.Parts {
			tainted = g.readPart(seen, message.Role, part) || tainted
		}
	}
	g.trust = trust{tainted: tainted, seen: seen, sources: g.trust.sources}
}

// readPart records the URLs one stored part shows the run, and reports
// whether it brought untrusted text into the context. The caller holds mu.
func (g *gate) readPart(seen seenURLs, role nacelle.Role, part nacelle.Part) bool {
	switch part := part.(type) {
	case nacelle.ToolCall:
		return g.sourceLocked(part.Name)
	case nacelle.ToolResult:
		if !g.sourceLocked(part.Name) {
			return false
		}
		seen.addText(part.Result)
		return true
	case nacelle.Text:
		if _, attached := attachmentOf(part.Text); attached {
			return true
		}
		if role == nacelle.RoleUser {
			seen.addOwn(part.Text)
		}
	}
	return false
}

// hooks taint the run as soon as an untrusted source returns, before the
// model reads what it returned.
func (g *gate) hooks() map[nacelle.HookPoint][]nacelle.Hook {
	return map[nacelle.HookPoint][]nacelle.Hook{nacelle.AfterToolCall: {g.observe}}
}

func (g *gate) observe(_ context.Context, ev nacelle.HookEvent) nacelle.HookResult {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.sourceLocked(ev.Call.Name) {
		g.trust.tainted = true
		if g.trust.seen == nil {
			g.trust.seen = seenURLs{}
		}
		g.trust.seen.addText(ev.Result)
	}
	return nacelle.HookResult{}
}

// asks reports whether a call waits for the owner: an outward tool always;
// and, once the run is tainted, a shell command, any memory write, and a
// sonar_fetch or browser_navigate whose URL the run has not seen verbatim in
// the owner's messages or an untrusted source's result. A memory or artifact
// delete is Mycelium's to confirm, every time.
func (g *gate) asks(name string, input json.RawMessage) bool {
	if g.outward[name] {
		return true
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if !g.trust.tainted {
		return false
	}
	switch name {
	case browserNavigateTool, sonarFetchTool:
		return !g.trust.seen.has(input)
	case memoryWriteTool, shellTool:
		return true
	}
	return false
}
