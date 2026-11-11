package runner

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/FacileStudio/nacelle"
	"github.com/FacileStudio/nacelle/mcp/client"
)

// catalogThreshold is the tool count past which a run's suite tools go
// behind nacelle's catalog: every schema resident at once costs context and
// degrades tool choice well before a server run's seventy tools.
const catalogThreshold = 20

// callTool is the catalog's router: a call to it runs the suite tool its input
// names, and the loop routes it before the gate, so approvals, taint and scope
// policies see that tool's own name and arguments.
const callTool = "call_tool"

// directSuiteTools stay mounted when the rest go behind the catalog. They
// are the tools the standing prompt tells the model to use unprompted (the
// web note's search and fetch, the memory note's search, read and save, the
// mail note's account list, search, thread, send and organize, the calendar
// note's list and create) and the reads most requests start from.
var directSuiteTools = []string{
	"sonar_search", "sonar_fetch",
	"memory_search", "memory_read", "memory_write",
	"courrier_list_accounts", "courrier_search_mail", "courrier_read_thread",
	"courrier_draft", "courrier_send", "courrier_organize",
	"agenda_list_events", "agenda_create_event", "opus_search_tasks", "glouton_contacts_search",
}

// mountSuite answers the suite tools as a run holding others more tools
// mounts them, and the prompt section that says how to reach the ones behind
// the catalog. A run small enough to carry every schema mounts them whole.
func mountSuite(tools []nacelle.Tool, others int) ([]nacelle.Tool, string) {
	if others+len(tools) <= catalogThreshold {
		return tools, ""
	}
	var direct, deferred []nacelle.Tool
	var names []string
	for _, tool := range tools {
		if slices.Contains(directSuiteTools, tool.Spec().Name) {
			direct = append(direct, tool)
			continue
		}
		deferred = append(deferred, tool)
		names = append(names, tool.Spec().Name)
	}
	if len(deferred) == 0 {
		return tools, ""
	}
	return append(direct, client.Catalog(deferred, 0)...), catalogNote(names)
}

// catalogNote names the app tools behind the catalog and how to run one.
func catalogNote(names []string) string {
	return "## App tools on demand\n\nTo keep this prompt short, these app tools are not mounted directly: " +
		strings.Join(names, ", ") + ". Run one with call_tool, passing its exact name and its arguments object. " +
		"Read its schema first with get_tool_details, and use search_tools when you are not sure which tool fits. " +
		"A tool reached this way works exactly as if it were mounted: the user sees the same approval card, under the tool's own name, before it changes anything."
}

// routed is the tool a call_tool call runs and the arguments it passes, so
// the session views show the app tool rather than the router. Any other call
// comes back as it is.
func routed(name string, input json.RawMessage) (string, json.RawMessage) {
	if name != callTool {
		return name, input
	}
	var in struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if json.Unmarshal(input, &in) != nil || in.Name == "" {
		return name, input
	}
	return in.Name, in.Arguments
}
