package tui

import (
	"github.com/FacileStudio/Bulle/apps/cli/internal/approval"
	"github.com/FacileStudio/Bulle/core/engine"
	"github.com/FacileStudio/Bulle/core/settings"
	"github.com/FacileStudio/nacelle"
)

const statusDone = "completed"

// BuildApprovals constructs the approval gate and returns the approval function.
func BuildApprovals(config settings.Config) (*approval.Approvals, nacelle.Approve) {
	return approval.Build(*config.ApproveTools)
}

// record keeps the conversation the run commits: every KindMessage, in
// order, so reasoning, tool images and provider Native parts replay. A tool
// result flushes the text the model streamed before asking for the tool, so
// it reaches the scrollback above the result.
func (m *Model) record(event nacelle.Event) {
	switch event.Kind {
	case nacelle.KindMessage:
		m.conversation = append(m.conversation, *event.Message)
	case nacelle.KindToolResult:
		if !event.Tool.Discarded {
			m.flush()
		}
	}
}

// keepUnfinished flushes what the run still holds and, when the run was cut
// off before it committed the turn the reader watched stream, keeps that text
// with why the run stopped. The conversation is then repaired: a run stopped
// mid-tool or capped leaves calls no result answers.
func (m *Model) keepUnfinished(stop nacelle.Stop) {
	said := m.flush()
	last := len(m.conversation) - 1
	if said != "" && (last < 0 || m.conversation[last].Role != nacelle.RoleAssistant) {
		parts := []nacelle.Part{nacelle.Text{Text: said}}
		if stop != "" {
			parts = append(parts, nacelle.Finish{Stop: stop})
		}
		m.conversation = append(m.conversation, nacelle.Message{Role: nacelle.RoleAssistant, Parts: parts})
	}
	m.conversation = engine.Repair(m.conversation)
}

func (m *Model) editing() int {
	i := m.hist.Editing(m.Len())
	if i < 0 || m.prompt.Value() != "" {
		return i
	}
	m.hist.FromEnd = 0
	return -1
}
