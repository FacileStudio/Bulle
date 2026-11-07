package tui

import (
	"strconv"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// scrolledModel is a tui-mode model holding n numbered transcript rows.
func scrolledModel(n int) *Model {
	m := tuiModel()
	for i := range n {
		m.hold = append(m.hold, "row"+strconv.Itoa(i))
	}
	return m
}

// The offset clamps to the top, so scrolling past it and back down moves at
// once rather than through rows that are not there.
func TestScrollClampsToTheTop(t *testing.T) {
	m := scrolledModel(50)
	for range 100 {
		m.scrollWheel(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	}
	m.tuiUpper(10)
	if want := m.scrollLen - 10; m.scrollTop != want {
		t.Errorf("scrollTop = %d, want clamped to %d", m.scrollTop, want)
	}
}

// A scrolled-back window stays on the rows it shows while new rows arrive.
func TestScrolledViewPinsToItsRows(t *testing.T) {
	m := scrolledModel(50)
	m.tuiUpper(10)
	m.scrollTop = 5
	before := m.tuiUpper(10)
	m.hold = append(m.hold, "new1", "new2", "new3")
	after := m.tuiUpper(10)
	if before[0] != after[0] {
		t.Errorf("window moved from %q to %q", before[0], after[0])
	}
}

// PgUp, PgDn, Home and End move the window.
func TestScrollKeys(t *testing.T) {
	m := scrolledModel(200)
	m.tuiUpper(10)
	m.Update(tea.KeyPressMsg{Code: tea.KeyPgUp})
	if m.scrollTop == 0 {
		t.Fatal("pgup did not scroll")
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnd})
	if m.scrollTop != 0 {
		t.Errorf("end left scrollTop at %d", m.scrollTop)
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyHome})
	if top := m.tuiUpper(10)[0]; top != m.hold[0] {
		t.Errorf("home shows %q first, want the top row %q", top, m.hold[0])
	}
}
