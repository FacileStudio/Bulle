package tui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

// reflowDelay is how long a resize waits for the next one before it repaints
// the held transcript, so dragging a window edge repaints once, not per step.
const reflowDelay = 50 * time.Millisecond

// reflowTick is a debounced reflow coming due; seq names the resize that
// scheduled it, and only the newest one repaints.
type reflowTick struct{ seq int }

// entryPaint is a held entry's last paint and the frame it was painted for.
type entryPaint struct {
	key paintKey
	out string
}

// scheduleReflow repaints the hold after the resize settles, or at once when
// no delay is set, which keeps model tests synchronous.
func (m *Model) scheduleReflow() tea.Cmd {
	if m.live.reflowGap == 0 {
		m.reflowHold()
		return nil
	}
	m.live.reflowSeq++
	seq := m.live.reflowSeq
	return tea.Tick(m.live.reflowGap, func(time.Time) tea.Msg { return reflowTick{seq: seq} })
}

// reflowDue repaints the hold if no later resize has superseded this one.
func (m *Model) reflowDue(tick reflowTick) tea.Cmd {
	if tick.seq == m.live.reflowSeq {
		m.reflowHold()
	}
	return nil
}

// cachedEntry paints held entry i at width, reusing its last paint when the
// width, renderer and backdrop are the same.
func (m *Model) cachedEntry(i, width int) string {
	entry := &m.held[i]
	key := m.paintKey("", width)
	if entry.paint.out == "" || entry.paint.key != key {
		entry.paint = entryPaint{key: key, out: strings.Trim(m.paintEntry(*entry, width), "\n")}
	}
	return entry.paint.out
}
