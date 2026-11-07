package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/FacileStudio/Bulle/core/compaction"
)

func (m *Model) softPass() tea.Cmd {
	m.takeStep(m.engine().SoftPass(m.conversation, m.size))
	return nil
}

func (m *Model) compactTiered(ctx context.Context) tea.Cmd {
	switch m.engine().Policy.Tier(m.size) {
	case compaction.Soft:
		return m.softPass()
	case compaction.Smart:
		return m.beginCompaction(ctx, false)
	default:
		return nil
	}
}

func (m *Model) thrashed() bool {
	return m.engine().Thrashed()
}

// maskReport is the card for a stub pass, with the kept-tail hint when the
// history was too small for a fold to land the conversation under its trigger.
func maskReport(outcome compaction.Outcome, last compacted) string {
	card := compactReport(outcome, last)
	if outcome.MaskOnly {
		card += "\n   cost sits in the kept tail — compaction protects the newest turns; /clear or read in chunks"
	}
	return card
}
