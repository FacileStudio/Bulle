package tui

import (
	"strings"

	"github.com/FacileStudio/Bulle/apps/cli/internal/toolview"
	"github.com/FacileStudio/nacelle"
)

// waitingNote replaces a call's clock while its approval is on screen.
const waitingNote = " · waiting for approval"

// deniedSuffix is how nacelle words a call the approval gate refused.
const deniedSuffix = "was not approved to run"

// denied reports whether a call failed only because the reader refused it.
func denied(err error) bool {
	return err != nil && strings.HasSuffix(err.Error(), deniedSuffix)
}

// sayDenied commits a refused call as a quiet "denied" line rather than a red
// failure carrying nacelle's error text.
func (m *Model) sayDenied(line string, tool *nacelle.ToolEvent) {
	m.flushFailures()
	m.say(fromTool, line+" · denied")
	m.drainEdit(tool.ID)
	delete(m.run.outputs, tool.ID)
}

// awaiting reports whether the shown approval is for this running call.
func (m *Model) awaiting(g toolview.Group) bool {
	p := m.run.pending
	return p != nil && g.Count <= 1 && g.Name == p.Name && g.Tool.Input == string(p.Input)
}

// flightLine is a running call's live line: its clock, or a note that it is
// waiting on the reader while its approval is shown.
func (m *Model) flightLine(g toolview.Group) string {
	if !m.awaiting(g) {
		return g.InFlightLine(m.width)
	}
	return g.GroupLine(m.width-len(waitingNote)) + waitingNote
}

// countFinished keeps a finished call's output for its box and counts it, a
// refusal not counting as a failure.
func (m *Model) countFinished(tool *nacelle.ToolEvent) {
	if m.run.outputs == nil {
		m.run.outputs = make(map[string]string)
	}
	if _, ok := m.run.outputs[tool.ID]; !ok {
		m.run.outputs[tool.ID] = tool.Result
	}
	m.tools++
	m.session.Tool(tool.Name, tool.Duration)
	if tool.Err != nil && !denied(tool.Err) {
		m.failed++
	}
}
