package runner

import (
	"slices"

	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/nacelle"
)

// serverConfirmed are the suite writes whose own server asks the owner before
// it acts, so a tainted run does not put a second card in front of them.
// Each was read in its server's MCP code: a tool goes here only once its
// handler is seen asking (tronc's Confirmations.Request or confirm.request).
// Docs confirms docs_replace alone and Feuilles confirms nothing, whatever
// their tool notes say.
var serverConfirmed = []string{
	"courrier_send", "courrier_delete_draft",
	"agenda_delete_event",
	"nuage_delete", "nuage_create_share",
	"opus_delete_task", "opus_delete_project",
	"sablier_delete_entry",
	"glouton_company_delete", "glouton_person_delete",
	"plume_send_document", "plume_remind_signer", "plume_delete_document",
	"capsule_revoke",
	"journal_delete_saved_query", "journal_delete_alert",
	"antenne_replay_event",
	"vision_delete_site", "vision_share_site", "vision_delete_goal",
	"echo_delete_room",
	"toile_delete_shapes",
	"docs_replace",
}

// readVerbs name a suite tool that only reads. A listed tool with none of
// them in its name counts as a write.
var readVerbs = []string{"list", "read", "get", "search", "find", "count", "status", "summary"}

var listedTools = func() map[string]bool {
	listed := map[string]bool{}
	for _, tools := range suite.MCPToolLists() {
		for _, name := range tools {
			listed[name] = true
		}
	}
	return listed
}()

func serverConfirms(name string) bool {
	return slices.Contains(serverConfirmed, name)
}

// readName reports whether a suite tool's name, past its app prefix, holds a
// read verb.
func readName(name string) bool {
	return hasVerb(name, readVerbs)
}

// openWorldSources names the mounted tools whose server says they read
// beyond its own system: read-only and open world. Bulle mounts first-party
// suite servers only, so it trusts their annotations for this.
func openWorldSources(tools []nacelle.Tool) []string {
	var names []string
	for _, tool := range tools {
		spec := tool.Spec()
		if spec.ReadOnly && spec.Hints != nil && spec.Hints.OpenWorld {
			names = append(names, spec.Name)
		}
	}
	return names
}

// mountedWrites names the mounted suite tools that are not read-only.
func mountedWrites(tools []nacelle.Tool) []string {
	var names []string
	for _, tool := range tools {
		if spec := tool.Spec(); !spec.ReadOnly {
			names = append(names, spec.Name)
		}
	}
	return names
}

// addSources makes names untrusted sources for the rest of the run.
func (g *gate) addSources(names []string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.trust.sources == nil {
		g.trust.sources = map[string]bool{}
	}
	for _, name := range names {
		g.trust.sources[name] = true
	}
}

// addWrites makes names suite writes for the rest of the run.
func (g *gate) addWrites(names []string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.trust.writes == nil {
		g.trust.writes = map[string]bool{}
	}
	for _, name := range names {
		g.trust.writes[name] = true
	}
}

// sourceLocked reports whether name is an untrusted source, listed or
// mounted. The caller holds mu.
func (g *gate) sourceLocked(name string) bool {
	return g.trust.sources[name] || untrusted(name)
}

// writeLocked reports whether name is a suite tool that changes something:
// mounted without a read-only annotation, or listed by Bulle without a read
// verb. The caller holds mu.
func (g *gate) writeLocked(name string) bool {
	return g.trust.writes[name] || (listedTools[name] && !readName(name))
}
