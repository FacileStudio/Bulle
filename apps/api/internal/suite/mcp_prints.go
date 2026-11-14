package suite

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"slices"

	"github.com/FacileStudio/nacelle"
)

// fingerprint is what a tool promises the model and the gate: its name, its
// input schema and its annotations. The description is left out, since a
// reworded sentence changes neither.
func fingerprint(spec nacelle.Spec) string {
	raw, err := json.Marshal(struct {
		Name     string         `json:"name"`
		Schema   map[string]any `json:"schema"`
		ReadOnly bool           `json:"read_only"`
		Hints    *nacelle.Hints `json:"hints"`
	}{spec.Name, spec.Schema, spec.ReadOnly, spec.Hints})
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

// watch compares each listed tool with what the process saw before. A tool
// whose name, schema or annotations changed is logged once per change, and
// a tool none of Bulle's lists names is logged once: the gate classifies
// tools by name, so either can quietly drop a taint source.
func (m *MCP) watch(app string, specs []nacelle.Spec) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, spec := range specs {
		got := fingerprint(spec)
		seen, known := m.prints[spec.Name]
		m.prints[spec.Name] = got
		switch {
		case !known && !listedTool(app, spec.Name):
			m.logger.Warn("mcp tool unknown to Bulle", "app", app, "tool", spec.Name)
		case known && seen != got:
			m.logger.Warn("mcp tool definition changed", "app", app, "tool", spec.Name)
		}
	}
}

func listedTool(app, name string) bool {
	tools, ok := MCPToolLists()[app]
	return !ok || bookkeepingTools[name] || slices.Contains(tools, name)
}

// MCPToolLists is every MCP app's tools as Bulle knows them, by app. The
// gate's source list and the test fakes follow it; a live tool missing from
// it is logged when listed.
func MCPToolLists() map[string][]string {
	return map[string][]string{
		AppCourrier: CourrierTools(), AppAgenda: AgendaTools(), AppSonar: SonarTools(), AppNuage: NuageTools(),
		AppOpus: OpusTools(), AppSablier: SablierTools(), AppGlouton: GloutonTools(), AppPlume: PlumeTools(),
		AppCapsule: CapsuleTools(), AppJournal: JournalTools(), AppAntenne: AntenneTools(), AppVision: VisionTools(),
		AppEcho: EchoTools(), AppToile: ToileTools(), AppDocs: DocsTools(), AppFeuilles: FeuillesTools(),
	}
}
