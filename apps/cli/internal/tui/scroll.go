package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// scrollAll is a scroll distance past any transcript, clamped to the top on the
// next frame.
const scrollAll = 1 << 30

// pinScroll keeps a scrolled-back window on the rows it shows while new rows
// arrive under it, and clamps the offset to what the content can scroll. At
// the newest row nothing is pinned, so the view follows the stream.
func (m *Model) pinScroll(total, avail int) {
	if m.scrollTop > 0 && m.scrollLen > 0 && total > m.scrollLen {
		m.scrollTop += total - m.scrollLen
	}
	m.scrollLen = total
	m.scrollTop = min(max(m.scrollTop, 0), max(total-avail, 0))
}

// scrollKey moves the tui-mode window by a page with PgUp and PgDn, and to the
// top or the newest row with Home and End while the prompt is empty, so the
// keys still move the cursor in typed text.
func (m *Model) scrollKey(press tea.KeyPressMsg) bool {
	if m.mode != modeTUI {
		return false
	}
	page := max(m.liveRows-2, 1)
	empty := strings.TrimSpace(m.prompt.Value()) == ""
	switch press.String() {
	case "pgup":
		m.scrollTop += page
	case "pgdown":
		m.scrollTop = max(m.scrollTop-page, 0)
	case "home":
		if !empty {
			return false
		}
		m.scrollTop = scrollAll
	case "end":
		if !empty {
			return false
		}
		m.scrollTop = 0
	default:
		return false
	}
	return true
}
