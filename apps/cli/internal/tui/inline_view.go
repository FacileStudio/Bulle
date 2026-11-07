package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// assembleInline is the scrollback render: everything already said lives in the
// terminal's own history, and the view draws only the live region, the prompt,
// and the menu above the prompt. The prompt is separated from the content above and
// below by one blank row each, so it reads as its own band. aboveContent
// already ends in a blank row when the menu is closed, so the separator is only
// added when the content does not already breathe.
//
// The cursor row is counted from the parts actually joined above the prompt
// bar, plus the bar's own padding row, so a separator that was or was not
// added never puts the cursor a row off.
func (m *Model) assembleInline() tea.View {
	parts := m.aboveContent()
	if last := len(parts) - 1; last < 0 || parts[last] != "" {
		parts = append(parts, "")
	}
	if menu := m.viewMenu(); menu != "" {
		parts = append(parts, menu)
	}
	promptRow := lipgloss.Height(strings.Join(parts, "\n")) + 1
	parts = append(parts, m.promptBar(highlightSkills(m.prompt.View(), m.skillTokens())))
	if below := m.belowContent(); below != "" {
		parts = append(parts, "", below)
	}
	body := strings.Join(parts, "\n")
	m.frameRows = lipgloss.Height(body)

	view := tea.NewView(body)
	if position := m.prompt.Cursor(); position != nil {
		position.Y += promptRow
		view.Cursor = position
	}
	return view
}
