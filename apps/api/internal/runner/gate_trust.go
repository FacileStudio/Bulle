package runner

import (
	"context"
	"encoding/json"
	"path"
	"slices"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/nacelle"
)

const (
	browserNavigateTool = "browser_navigate"
	memoryWriteTool     = "memory_write"
	memoryDeleteTool    = "memory_delete"
)

// untrustedSources is every tool whose result carries text that somebody
// other than the owner wrote: mail, the web, the browser, shared files, tasks
// and time, and prospects imported or audited from the web. A
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
}

// shellTool runs commands in the computer, which has open egress.
const shellTool = "bash"

// trust is what the gate knows about where the run's context came from:
// whether untrusted text reached it, and what the owner wrote themselves.
type trust struct {
	tainted  bool
	userText string
	sources  map[string]bool
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
// tool that ran, which is how a source reached through call_tool is caught. The owner's own text is kept for the
// browser_navigate check.
func (g *gate) seed(history []nacelle.Message) {
	var own strings.Builder
	tainted := false
	for _, message := range history {
		for _, part := range message.Parts {
			text, taints := readPart(message.Role, part)
			tainted = tainted || taints
			own.WriteString(text)
		}
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.trust = trust{tainted: tainted, userText: own.String(), sources: g.trust.sources}
}

// readPart is what one stored part adds to the owner's own text, and whether
// it brought untrusted text into the context.
func readPart(role nacelle.Role, part nacelle.Part) (string, bool) {
	switch part := part.(type) {
	case nacelle.ToolCall:
		return "", untrusted(part.Name)
	case nacelle.ToolResult:
		return "", untrusted(part.Name)
	case nacelle.Text:
		if _, attached := attachmentOf(part.Text); attached {
			return "", true
		}
		if role == nacelle.RoleUser {
			return part.Text + "\n", false
		}
	}
	return "", false
}

// hooks taint the run as soon as an untrusted source returns, before the
// model reads what it returned.
func (g *gate) hooks() map[nacelle.HookPoint][]nacelle.Hook {
	return map[nacelle.HookPoint][]nacelle.Hook{nacelle.AfterToolCall: {g.observe}}
}

func (g *gate) observe(_ context.Context, ev nacelle.HookEvent) nacelle.HookResult {
	if g.untrusted(ev.Call.Name) {
		g.mu.Lock()
		g.trust.tainted = true
		g.mu.Unlock()
	}
	return nacelle.HookResult{}
}

// asks reports whether a call waits for the owner: an outward tool always, a
// shell command, a profile.md write or a memory delete once the run is
// tainted, and a browser_navigate whose URL carries data the owner never
// wrote.
func (g *gate) asks(name string, input json.RawMessage) bool {
	if g.outward[name] {
		return true
	}
	g.mu.Lock()
	t := g.trust
	g.mu.Unlock()
	switch name {
	case browserNavigateTool:
		return unseenData(input, t.userText)
	case memoryWriteTool:
		return t.tainted && writesProfile(input)
	case shellTool, memoryDeleteTool:
		return t.tainted
	}
	return false
}

func writesProfile(input json.RawMessage) bool {
	var in struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(input, &in); err != nil {
		return true
	}
	return strings.EqualFold(path.Clean(strings.TrimSpace(in.Path)), suite.ProfilePage)
}
