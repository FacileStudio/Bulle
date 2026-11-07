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
)

// untrustedSources is every tool whose result carries text that somebody
// other than the owner wrote: mail, the web, the browser and Nuage files. A
// name ending in "_" covers every tool with that prefix. Old and new names
// both stay listed, so a rename never quietly drops a source.
var untrustedSources = []string{
	"courrier_inbox", "courrier_read", "courrier_search", "courrier_list_inbox", "courrier_get_email",
	"web_fetch", "web_search", "browser_",
	"files_list", "files_read", "files_search", "nuage_list", "nuage_read", "nuage_search",
}

// shellTool runs commands in the computer, which has open egress.
const shellTool = "bash"

// trust is what the gate knows about where the run's context came from:
// whether untrusted text reached it, and what the owner wrote themselves.
type trust struct {
	tainted  bool
	userText string
}

func untrusted(name string) bool {
	return slices.ContainsFunc(untrustedSources, func(source string) bool {
		if strings.HasSuffix(source, "_") {
			return strings.HasPrefix(name, source)
		}
		return name == source
	})
}

// seed reads the conversation the run starts from. A call of an untrusted
// source, or an attachment, anywhere in it taints the run, since that text is
// still in the model's context. The owner's own text is kept for the
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
	g.trust = trust{tainted: tainted, userText: own.String()}
}

// readPart is what one stored part adds to the owner's own text, and whether
// it brought untrusted text into the context.
func readPart(role nacelle.Role, part nacelle.Part) (string, bool) {
	switch part := part.(type) {
	case nacelle.ToolCall:
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
	if untrusted(ev.Call.Name) {
		g.mu.Lock()
		g.trust.tainted = true
		g.mu.Unlock()
	}
	return nacelle.HookResult{}
}

// asks reports whether a call waits for the owner: an outward tool always, a
// shell command or a profile.md write once the run is tainted, and a
// browser_navigate whose URL carries data the owner never wrote.
func (g *gate) asks(name string, input json.RawMessage) bool {
	if g.outward[name] {
		return true
	}
	g.mu.Lock()
	t := g.trust
	g.mu.Unlock()
	switch {
	case name == browserNavigateTool:
		return unseenData(input, t.userText)
	case name == memoryWriteTool:
		return t.tainted && writesProfile(input)
	case name == shellTool:
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
