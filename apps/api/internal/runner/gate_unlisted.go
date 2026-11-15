package runner

import "github.com/FacileStudio/nacelle"

// unlistedReads names the mounted read-only tools that none of Bulle's
// lists names, such as a tool an app shipped before Bulle learned of it. The
// gate cannot tell whether such a tool returns other people's text, so it
// counts as a source.
func unlistedReads(tools []nacelle.Tool) []string {
	var names []string
	for _, tool := range tools {
		if spec := tool.Spec(); spec.ReadOnly && !listedTools[spec.Name] {
			names = append(names, spec.Name)
		}
	}
	return names
}

// mountedSources names the mounted tools the run treats as sources: the open
// world reads and the reads Bulle does not list.
func mountedSources(tools []nacelle.Tool) []string {
	return append(openWorldSources(tools), unlistedReads(tools)...)
}
