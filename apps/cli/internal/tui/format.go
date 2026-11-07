package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/FacileStudio/Bulle/apps/cli/internal/theme"
	"github.com/FacileStudio/nacelle"
)

// question draws the reader's message as muse draws it in the browser: a
// full-width block on the surface tone with a bright spine (▌) on every row,
// one padded row above and below. The body renders at two columns short of
// the width, so each wrapped line keeps room for the spine and every row is
// padded to the same edge: the backdrop is a clean rectangle that never
// bleeds onto the row after it.
func (m *Model) question(text string, width int) string {
	inner := max(width-2, 1)
	block := m.theme.Question.Width(inner)
	spine := m.theme.Question.Foreground(m.theme.Bar.Text).Render("▌ ")
	tokens := m.skillTokens()
	lines := strings.Split(block.Render(text), "\n")
	rows := make([]string, 0, len(lines)+2)
	pad := spine + block.Render(strings.Repeat(" ", inner))
	rows = append(rows, pad)
	for _, line := range lines {
		rows = append(rows, spine+highlightSkills(line, tokens))
	}
	return strings.Join(append(rows, pad), "\n")
}

// thinking draws the model's reasoning as a quiet block: italic muted text
// behind a thin dim spine (│), lighter than the answer's ▌ so the two never
// read as one block.
func (m *Model) thinking(text string, width int) string {
	spine := theme.Fore(theme.Ink.Dim).Render("│ ")
	lines := strings.Split(m.theme.Thinking.Width(max(width-2, 1)).Render(text), "\n")
	for i, line := range lines {
		lines[i] = spine + line
	}
	return strings.Join(lines, "\n")
}

// gutterRows gives every row of a committed ⏺ or │ line the one-column
// gutter margined gives the same row while it is live, so committing a line
// never shifts it sideways.
func gutterRows(text string) string {
	return " " + strings.ReplaceAll(text, "\n", "\n ")
}

// wrapWidth is the width client lines wrap at, or 0 before the terminal has
// reported its size: the banner is said at the 80 columns the model starts
// with, and wrapping it there would break it early on a wider terminal.
func (m *Model) wrapWidth() int {
	if m.windowHeight == 0 {
		return 0
	}
	return max(m.width, 1)
}

// minWrap is the narrowest width a client line is wrapped at.
const minWrap = 10

// indented paints a client-side line in style, wrapped at width less the
// one-column indent every row carries, so a wrapped row lines up with the
// first. A width too narrow to hold a word leaves the text unwrapped.
func indented(style lipgloss.Style, text string, width int) string {
	if width > minWrap {
		text = ansi.Wrap(text, width-1, "")
	}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = " " + style.Render(line)
	}
	return strings.Join(lines, "\n")
}

// countedNoun formats a count with its English plural, singular for exactly
// one.
func countedNoun(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

func cutShort(stop nacelle.Stop) string {
	if stop == "" || stop.Complete() {
		return ""
	}
	switch stop {
	case nacelle.StopMaxTokens:
		return "cut off at the token limit"
	case nacelle.StopContext:
		return "cut off: out of context"
	case nacelle.StopRefusal:
		return "refused by the model"
	case nacelle.StopIterations:
		return "stopped at the iteration limit"
	case abandoned:
		return "abandoned"
	}
	return "stopped early"
}
