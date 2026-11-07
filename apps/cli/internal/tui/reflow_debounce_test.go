package tui

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// A resize under a debounce waits for its tick, and only the newest resize's
// tick repaints the hold.
func TestResizeReflowWaitsForTheLastResize(t *testing.T) {
	m := tuiModel()
	m.live.reflowGap = time.Hour
	m.say(fromModel, strings.Repeat("words that wrap ", 12))
	m.prints()

	first := m.resize(tea.WindowSizeMsg{Width: 40, Height: 24})
	m.resize(tea.WindowSizeMsg{Width: 30, Height: 24})
	if first == nil {
		t.Fatal("a debounced resize scheduled nothing")
	}
	m.reflowDue(reflowTick{seq: m.live.reflowSeq - 1})
	if widest(m.hold) <= 30 {
		t.Fatal("a superseded tick repainted the hold")
	}
	m.reflowDue(reflowTick{seq: m.live.reflowSeq})
	if w := widest(m.hold); w > 30 {
		t.Errorf("hold is %d wide after the last tick, want <= 30", w)
	}
}

// A held entry repainted at an unchanged width reuses its paint.
func TestHeldEntryPaintIsCached(t *testing.T) {
	m := tuiModel()
	m.say(fromModel, "cached **answer**")
	m.renderHeld(m.width)
	last := len(m.held) - 1
	m.held[last].paint.out = "sentinel"
	if rows := strings.Join(m.renderHeld(m.width), "\n"); !strings.Contains(rows, "sentinel") {
		t.Error("an unchanged width repainted the entry")
	}
	if rows := strings.Join(m.renderHeld(m.width-10), "\n"); strings.Contains(rows, "sentinel") {
		t.Error("a new width reused the old paint")
	}
}

// widest is the widest row in rows.
func widest(rows []string) int {
	w := 0
	for _, row := range rows {
		w = max(w, lipgloss.Width(row))
	}
	return w
}
