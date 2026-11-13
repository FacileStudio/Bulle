package runner

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/FacileStudio/nacelle"
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
// the catalog. A run small enough to carry every schema mounts them whole,
// unless an app did not answer: the catalog is then mounted too, so its
// tools can join the run once it answers again.
func mountSuite(tools []nacelle.Tool, others int, absent absentApps) ([]nacelle.Tool, string) {
	small := others+len(tools) <= catalogThreshold
	down := len(downApps(absent.apps)) > 0
	if small && !down {
		return tools, ""
	}
	var direct []nacelle.Tool
	var names []string
	for _, tool := range tools {
		if small || slices.Contains(directSuiteTools, tool.Spec().Name) {
			direct = append(direct, tool)
			continue
		}
		names = append(names, tool.Spec().Name)
	}
	if len(names) == 0 && !down {
		return tools, ""
	}
	return append(direct, newSuiteCatalog(tools, absent).tools()...), catalogNote(names)
}

// catalogNote names the app tools behind the catalog and how to run one.
func catalogNote(names []string) string {
	if len(names) == 0 {
		return "## App tools on demand\n\nsearch_tools finds the app tools that join this conversation after it started; " +
			"run one with call_tool, passing its exact name and its arguments object, after reading its schema with get_tool_details."
	}
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

// found lists the matches: every name when the query was empty, else the
// first searchLimit with their descriptions and a line when more matched.
func found(specs []nacelle.Spec, namesOnly bool) string {
	var body strings.Builder
	if namesOnly {
		for _, spec := range specs {
			fmt.Fprintln(&body, spec.Name)
		}
		return body.String()
	}
	for _, spec := range specs[:min(len(specs), searchLimit)] {
		fmt.Fprintf(&body, "%s — %s\n", spec.Name, spec.Description)
	}
	if len(specs) > searchLimit {
		fmt.Fprintf(&body, "[%d more tools match; add words to narrow the search]\n", len(specs)-searchLimit)
	}
	return body.String()
}
