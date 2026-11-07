package settings

// Toggles is the tool-mount settings: whether the model gets each optional
// tool. Every toggle is a pointer so "not in this file" can be told from "false".
type Toggles struct {
	Bash           *bool         `yaml:"bash"`
	ParallelAgents *bool         `yaml:"parallel_agents"`
	Tasks          *bool         `yaml:"todo_write"`
	Diagnostics    *bool         `yaml:"diagnostics"`
	SearchContent  *bool         `yaml:"grep"`
	FindFiles      *bool         `yaml:"glob"`
	Legacy         LegacyToggles `yaml:",inline"`
	// GitCoauthor makes every commit the agent creates carry its
	// "Co-authored-by: Bulle Agent <bulle@facile.studio>" trailer; unset
	// means on.
	GitCoauthor *bool `yaml:"git_coauthor"`
}

// LegacyToggles reads the tool keys a config written before nacelle v0.32
// used. decodeConfig folds them into the current keys, which win when a file
// carries both.
type LegacyToggles struct {
	RunCommand    *bool `yaml:"run_command"`
	Tasks         *bool `yaml:"tasks"`
	SearchContent *bool `yaml:"search_content"`
	FindFiles     *bool `yaml:"find_files"`
}

// adoptLegacy moves every old key the file set onto its current one.
func (t *Toggles) adoptLegacy() {
	for _, pair := range []struct{ current, old **bool }{
		{&t.Bash, &t.Legacy.RunCommand},
		{&t.Tasks, &t.Legacy.Tasks},
		{&t.SearchContent, &t.Legacy.SearchContent},
		{&t.FindFiles, &t.Legacy.FindFiles},
	} {
		if *pair.current == nil {
			*pair.current = *pair.old
		}
	}
	t.Legacy = LegacyToggles{}
}
