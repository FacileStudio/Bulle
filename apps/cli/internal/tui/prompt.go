package tui

import (
	"strings"

	"github.com/FacileStudio/Bulle/apps/cli/internal/theme"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const promptRows = 10

// maxContentRows is the prompt's input gate, far above anything a message
// holds. Bubbles blocks newline insertion once content passes the gate —
// with MaxHeight doing that job, a paste of ten-plus lines silently ate
// every later Alt+Enter — so the gate lives on MaxContentHeight while
// MaxHeight stays the visible viewport cap the prompt scrolls inside.
const maxContentRows = 1000

// minHeightRows is the shortest the input renders even when it holds a single
// line, so the field reads as a single line rather than a padded bar. The blank
// row above the whole field is added by the view assembly.
const minHeightRows = 1

// newPrompt builds the compose textarea on the flat input panel (see
// promptBar). placeholder is the ghost text shown while the prompt is empty.
// The gutter holding the label is set by the model, which owns the run state
// the label shows.
func newPrompt(placeholder string, bar theme.Bar) textarea.Model {
	prompt := textarea.New()
	prompt.Placeholder = placeholder
	prompt.ShowLineNumbers = false
	prompt.DynamicHeight = true
	prompt.MinHeight = minHeightRows
	prompt.MaxHeight = promptRows
	prompt.MaxContentHeight = maxContentRows
	prompt.KeyMap.InsertNewline = key.NewBinding(key.WithKeys("alt+enter", "shift+enter"))
	prompt.SetVirtualCursor(false)
	stylePrompt(&prompt, bar)
	prompt.Focus()
	return prompt
}

// stylePrompt paints every row, focused or not, on the panel colour, with
// text in the bar's text colour and the placeholder muted. Base is inherited
// by every computed style, which is what puts the panel under all rows; the
// cursor line gets the same ground so the row being typed on does not stand
// out as a stripe.
func stylePrompt(prompt *textarea.Model, bar theme.Bar) {
	ground := lipgloss.NewStyle().Background(bar.Panel).Foreground(bar.Text)
	styles := prompt.Styles()
	for _, state := range []*textarea.StyleState{&styles.Focused, &styles.Blurred} {
		state.Base = state.Base.Inherit(ground)
		state.CursorLine = ground
		state.Text = ground
		state.Placeholder = lipgloss.NewStyle().Background(bar.Panel).Foreground(bar.Muted)
		state.EndOfBuffer = ground
	}
	prompt.SetStyles(styles)
}

// ask takes the line the reader typed and submits it. The prompt is cleared
// first, so the field is empty whether the line goes out now or is queued.
func (m *Model) ask() tea.Cmd {
	if strings.TrimSpace(m.prompt.Value()) == "" {
		return nil
	}
	question := m.takePrompt()
	m.prompt.Reset()
	return m.submit(question)
}

// submit sends one line as a prompt, or queues it behind the run in flight.
//
// It is the one path a question takes, whether it was typed at the prompt or
// sent by an attached editor: a line arriving while a run is going is queued
// rather than dropped, and the choice between the two lives here rather than in
// each caller.
func (m *Model) submit(question string) tea.Cmd {
	question = strings.TrimSpace(question)
	if question == "" {
		return nil
	}
	if !m.run.busy && !m.compacting {
		m.hist.Remember(question, m.Items())
		m.layout(m.windowHeight)
		return m.dispatch(question)
	}

	held := m.hist.Requeue(m.Items(), question)
	if !held {
		m.Add(question)
	}
	m.hist.Remember(question, m.Items())
	m.layout(m.windowHeight)
	return nil
}

func (m *Model) dispatch(line string) tea.Cmd {
	m.grind.used = 0
	m.say(fromReader, echo(line))

	started := tea.Cmd(nil)
	if cmd, ok := m.parseCommand(line); ok {
		started = cmd(m)
	} else {
		started = m.send(line)
	}
	return tea.Sequence(m.prints(), started)
}
