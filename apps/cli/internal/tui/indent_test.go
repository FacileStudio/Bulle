package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Client, turn and compact lines wrap at the width and every wrapped row keeps
// the one-column indent, not only the first.
func TestClientLinesIndentEveryWrappedRow(t *testing.T) {
	m := sized()
	m.resize(tea.WindowSizeMsg{Width: 30, Height: 24})
	long := strings.Repeat("word ", 20)
	for _, who := range []speaker{fromClient, fromTurn, fromCompact} {
		rows := strings.Split(m.paint(who, long), "\n")
		if len(rows) < 2 {
			t.Fatalf("speaker %d did not wrap: %q", who, rows)
		}
		for _, row := range rows {
			if !strings.HasPrefix(visible(row), " ") || lipgloss.Width(row) > 30 {
				t.Errorf("speaker %d row %q lacks the indent or overflows", who, visible(row))
			}
		}
	}
}

// Before the terminal reports its size a client line is not wrapped, so the
// banner is not broken at the 80 columns the model starts with.
func TestClientLinesWaitForTheWindowBeforeWrapping(t *testing.T) {
	m := bareBanner()
	if rows := strings.Split(m.paint(fromClient, strings.Repeat("word ", 30)), "\n"); len(rows) != 1 {
		t.Errorf("unsized client line wrapped into %d rows", len(rows))
	}
}
