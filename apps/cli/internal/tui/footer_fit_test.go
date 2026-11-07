package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// At a narrow width the keys stay whole and the left side gives up its items
// from the end, so the footer never overflows or cuts a key in half.
func TestFooterKeepsTheKeysAndDropsLeftItems(t *testing.T) {
	m := sized()
	m.activeModel = "anthropic/claude-haiku-4.5-with-a-long-name"
	m.resize(tea.WindowSizeMsg{Width: 60, Height: 24})
	foot := visible(m.barFooter())
	if w := lipgloss.Width(foot); w > 60 {
		t.Errorf("footer is %d wide at 60 columns: %q", w, foot)
	}
	if !strings.HasSuffix(strings.TrimRight(foot, " "), "ctrl+g editor") {
		t.Errorf("footer lost its keys: %q", foot)
	}
}
