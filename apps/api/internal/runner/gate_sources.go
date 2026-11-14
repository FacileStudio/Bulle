package runner

import "github.com/FacileStudio/nacelle"

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

// sourceLocked reports whether name is an untrusted source, listed or
// mounted. The caller holds mu.
func (g *gate) sourceLocked(name string) bool {
	return g.trust.sources[name] || untrusted(name)
}
