package tui

import (
	"strings"
	"time"

	"charm.land/glamour/v2"
)

// liveFrameGap is how long a streamed render may go stale while deltas keep
// arriving: about 30 renders a second, however fast the model writes.
const liveFrameGap = 33 * time.Millisecond

// fenceSlack is how many source lines past the live region an open fence keeps,
// so the rows the reader can see are rendered with their context.
const fenceSlack = 20

// paintKey is everything a live render depends on. pretty changes on every
// width or theme change, because restyle rebuilds the renderer.
type paintKey struct {
	text        string
	width       int
	pretty      *glamour.TermRenderer
	transparent bool
}

// cachedPaint is one live block's last render and when it was drawn.
type cachedPaint struct {
	key paintKey
	out string
	at  time.Time
}

// liveCache holds the streaming region's renders between frames, so a spinner
// tick costs nothing and a fast stream re-renders at most once per gap. gap is
// zero until the program starts, which keeps model tests fresh on every frame.
type liveCache struct {
	answer  cachedPaint
	thought cachedPaint
	rows    marginedRows
	gap     time.Duration
}

// marginedRows is the streaming region's last margined rows, keyed on the
// joined rows and the width they were fitted to.
type marginedRows struct {
	key   string
	width int
	out   []string
}

// marginedLive is margined for the streaming rows, cached, because fitting
// every row of a long answer to the width is most of an unchanged frame.
func (m *Model) marginedLive(live []string) []string {
	key := strings.Join(live, "\n")
	if m.live.rows.out == nil || m.live.rows.key != key || m.live.rows.width != m.width {
		m.live.rows = marginedRows{key: key, width: m.width, out: m.margined(live)}
	}
	return m.live.rows.out
}

// paintKey builds the cache key for text at width.
func (m *Model) paintKey(text string, width int) paintKey {
	return paintKey{text: text, width: width, pretty: m.pretty, transparent: m.transparent}
}

// cachedRender returns slot's render for key, painting it only when the key
// changed. A change in the text alone keeps the old render until the gap has
// passed since it was drawn.
func (m *Model) cachedRender(slot *cachedPaint, key paintKey, paint func() string) string {
	if slot.out != "" && slot.key == key {
		return slot.out
	}
	sameFrame := slot.key.width == key.width && slot.key.pretty == key.pretty && slot.key.transparent == key.transparent
	if slot.out != "" && sameFrame && time.Since(slot.at) < m.live.gap {
		return slot.out
	}
	slot.key, slot.out, slot.at = key, paint(), time.Now()
	return slot.out
}

// liveAnswer is the streaming answer pane, cached, with an open fence cut to
// the tail the live region can show.
func (m *Model) liveAnswer(answer string) string {
	width := max(m.width, 1)
	text := fenceTail(answer, m.liveRows+fenceSlack)
	return m.cachedRender(&m.live.answer, m.paintKey(text, width), func() string { return m.answerPane(text, width) })
}

// liveThought is the streaming reasoning block, cached.
func (m *Model) liveThought(text string) string {
	width := max(m.width-2, 1)
	return m.cachedRender(&m.live.thought, m.paintKey(text, width), func() string { return m.thinking(text, width) })
}
