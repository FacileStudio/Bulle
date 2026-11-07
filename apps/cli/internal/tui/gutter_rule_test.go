package tui

import (
	"strings"
	"testing"
)

// Reasoning drawn live and the same reasoning committed land on the same
// columns, so committing it never shifts a row sideways.
func TestReasoningKeepsItsColumnsWhenCommitted(t *testing.T) {
	m := sized()
	text := strings.Repeat("considering the options ", 8)
	m.Expanded = true
	m.run.reasoning.WriteString(text)
	live := m.marginedLive(m.streaming())
	committed := strings.Split(m.paint(fromThinking, text), "\n")
	if len(live) != len(committed) {
		t.Fatalf("live has %d rows, committed %d", len(live), len(committed))
	}
	for i := range live {
		l, c := strings.TrimRight(visible(live[i]), " "), strings.TrimRight(visible(committed[i]), " ")
		if l != c {
			t.Errorf("row %d moved:\nlive      %q\ncommitted %q", i, l, c)
		}
	}
}

// A committed tool line carries the same one-column gutter the live row gets.
func TestToolLineCarriesTheLiveGutter(t *testing.T) {
	m := sized()
	if got := visible(m.paint(fromTool, "⏺ read_file(a.go)")); !strings.HasPrefix(got, " ⏺") {
		t.Errorf("committed tool line = %q, want the one-column gutter", got)
	}
}
