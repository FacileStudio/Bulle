package tui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/textarea"
	"charm.land/lipgloss/v2"

	"github.com/FacileStudio/Bulle/apps/cli/internal/layout"
	"github.com/FacileStudio/Bulle/apps/cli/internal/status"
	"github.com/FacileStudio/Bulle/core/compaction"
)

const (
	barLabel      = "Bulle"
	barLabelWidth = 5
	barPad        = 2
	barGutter     = barPad + barLabelWidth + barPad
)

var loaderFrames = []struct {
	active  int
	forward bool
}{{0, true}, {1, true}, {2, true}, {1, false}}

// gutter is the prompt's left column, after grok-cli: the label on the first
// row and blank panel on the others. While a run is going the label is the
// three-box loader, so the input says it is busy without a line of its own.
func (m *Model) gutter(info textarea.PromptInfo) string {
	bar := m.theme.Bar
	ground := lipgloss.NewStyle().Background(bar.Panel)
	if info.LineNumber != 0 {
		return ground.Render(strings.Repeat(" ", barGutter))
	}
	label := ground.Foreground(bar.Muted).Bold(true).Render(barLabel)
	if m.run.busy || m.compacting {
		label = m.loader() + ground.Render("  ")
	}
	return ground.Render(strings.Repeat(" ", barPad)) + label + ground.Render(strings.Repeat(" ", barPad))
}

func (m *Model) loader() string {
	bar := m.theme.Bar
	step := loaderFrames[int(time.Now().UnixMilli()/120)%len(loaderFrames)]
	var out strings.Builder
	for i := range 3 {
		distance := step.active - i
		if !step.forward {
			distance = i - step.active
		}
		glyph, shade := "⬝", bar.Dim
		switch {
		case distance == 0:
			glyph, shade = "■", bar.Text
		case distance == 1:
			glyph, shade = "■", bar.Muted
		}
		out.WriteString(lipgloss.NewStyle().Background(bar.Panel).Foreground(shade).Render(glyph))
	}
	return out.String()
}

// promptBar is the whole bottom block: the input on a flat panel with a row
// of padding above and below, then one line with the model and context on
// the left and the keys that work right now on the right.
func (m *Model) promptBar(prompt string) string {
	pad := lipgloss.NewStyle().Background(m.theme.Bar.Panel).Render(strings.Repeat(" ", max(m.width, 1)))
	return pad + "\n" + prompt + "\n" + pad + "\n" + m.barFooter()
}

// barRows is how many rows promptBar adds around the input itself.
const barRows = 3

func (m *Model) barFooter() string {
	bar := m.theme.Bar
	text := lipgloss.NewStyle().Foreground(bar.Text)
	muted := lipgloss.NewStyle().Foreground(bar.Muted)
	dim := lipgloss.NewStyle().Foreground(bar.Dim)

	left := text.Render(m.activeModel)
	if load := m.contextLeft(); load != "" {
		left += " " + load
	}
	if spent := m.spentLine(); spent != "" {
		left += " " + dim.Render(spent)
	}
	var keys []string
	if m.run.busy {
		keys = append(keys, muted.Render(layout.Unstyled(m.working())))
	}
	for _, k := range m.barKeys() {
		keys = append(keys, text.Render(k[0])+" "+muted.Render(k[1]))
	}
	right := strings.Join(keys, "  ")
	gap := max(m.width-2*barPad-lipgloss.Width(left)-lipgloss.Width(right), 1)
	line := strings.Repeat(" ", barPad) + left + strings.Repeat(" ", gap) + right
	return layout.Shave(line, max(m.width, 1))
}

func (m *Model) barKeys() [][2]string {
	switch {
	case m.run.pending != nil:
		return [][2]string{{"y", "once"}, {"a", "always"}, {"n", "deny"}}
	case m.run.busy:
		return [][2]string{{"enter", "queue"}, {"esc", "interrupt"}}
	case m.menu.Open():
		return [][2]string{{"tab", "accept"}, {"↑↓", "navigate"}, {"esc", "dismiss"}}
	}
	return [][2]string{{"/", "commands"}, {"shift+enter", "new line"}, {"ctrl+g", "editor"}}
}

// contextLeft is how much of the usable window is left, as grok shows it: a
// muted percentage and the remaining tokens dimmed.
func (m *Model) contextLeft() string {
	usable := m.engine().Policy.Usable()
	if usable <= 0 {
		return ""
	}
	size := m.size + m.run.liveOut
	left := max(usable-size, 0)
	muted := lipgloss.NewStyle().Foreground(m.theme.Bar.Muted)
	dim := lipgloss.NewStyle().Foreground(m.theme.Bar.Dim)
	text := muted.Render(fmt.Sprintf("%d%%", left*100/usable)) + " " + dim.Render(status.ShortTokens(left))
	if tier := m.engine().Policy.Tier(size); tier != compaction.Below {
		text += " " + dim.Render(tier.String())
	}
	return text
}

// spentLine is the session's price and tokens, and what is left of a grind
// budget, from the same figures as footer.
func (m *Model) spentLine() string {
	parts := m.footer()
	if len(parts) > 0 && m.activeModel != "" {
		parts = parts[1:]
	}
	var keep []string
	for _, p := range parts {
		if !strings.HasPrefix(p, "↕") {
			keep = append(keep, p)
		}
	}
	return strings.Join(keep, " · ")
}

// notice is the state line drawn above the input, and only when something
// needs the reader: an approval, a stop in progress, a run that ended short,
// or a session log that cannot be written. Ready says nothing and a running
// turn shows in the input's loader and the line under it, as in grok-cli.
func (m *Model) notice() string {
	state, attention := m.state()
	if !attention {
		return ""
	}
	return lipgloss.NewStyle().Foreground(m.theme.Bar.Text).Render(layout.Truncate(state, max(m.width, 1)))
}
