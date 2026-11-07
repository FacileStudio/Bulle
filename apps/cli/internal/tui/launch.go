package tui

import (
	"fmt"

	"github.com/FacileStudio/Bulle/apps/cli/internal/herdr"
	"github.com/FacileStudio/Bulle/core/sessions"
	"github.com/FacileStudio/Bulle/core/usage"
)

// boot wires the session config the model was built from into its live fields —
// the ones that arrive as pointers or are only meaningful on the running model
// rather than at construction. An attached editor is handed the model as its
// command handler here, because the handler is this session's prompt loop and
// this is the first moment it exists.
func boot(m *Model, c UISession) {
	m.groupTools = c.GroupTools != nil && *c.GroupTools
	m.Expanded = c.ShowThinking
	m.run.root = c.Root
	m.run.diffs = c.Diffs
	m.delegate = c.DelegateConfig
	m.activate(c.Backend, c.Model, c.BaseURL, c.APIKey)
	m.models = c.Models
	m.mode = renderMode(c.Mode)
	m.transparent = c.TransparentBlocks
	m.live.gap = liveFrameGap
	m.ide.surface = c.IDE
	if c.IDE != nil {
		c.IDE.SetCommands(m)
	}
	if c.Resume != "" || c.AutoResume {
		if res := sessions.RestoreAtLaunch(c.Resume, c.Root, c.AutoResume); res.Path != "" {
			m.session = sessions.OpenResumeSession(res.Path, c.Backend, c.Model, c.Root)
		}
	}
	if m.session == nil {
		m.session = sessions.OpenSession(c.Backend, c.Model, c.Root)
	}
	herdr.SetSession(m.herdrClient, m.session.Path())
}

// activate points the client at the backend it will bill from here, whether at
// launch or after a /model switch. It is the one place those five fields are
// written, so the two paths cannot leave the client naming one model and
// billing another; the usage sink follows the model for the same reason.
//
// The realised cost-per-token rate is dropped with it. That rate was measured
// on the old model's prices, and keeping it would scale the live price of the
// new model's tokens by them — a price nobody was ever charged.
func (m *Model) activate(backend, model, baseURL, apiKey string) {
	m.activeBackend, m.activeModel = backend, model
	m.activeBaseURL, m.activeAPIKey = baseURL, apiKey
	m.rate = 0
	m.sink = usage.NewSink(m.run.root, model)
}

// startupPrint hands the pre-queued banner lines to the terminal — unless it is
// running on the alternate screen, where Println would be a no-op and the lines
// are held to be drawn back into the view instead.
func startupPrint(m *Model) {
	if m.mode == modeTUI {
		return
	}
	for _, line := range m.unprinted {
		fmt.Println(line)
	}
	m.unprinted = nil
}
