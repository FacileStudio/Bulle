package tui

import (
	"fmt"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/FacileStudio/Bulle/core/askuser"
)

const (
	singleHint = "1-4 or ↑↓ enter to pick · or type your own answer and press enter · esc dismisses"
	multiHint  = "space or 1-4 toggles · enter confirms · type to add your own · esc dismisses"
)

// questionView is the pending question drawn above the prompt: its title,
// one numbered row per option, and the keys that answer it. Everything wraps
// at four columns short of the width, inside the live region's margin.
func (m *Model) questionView() string {
	q := m.run.question
	if q == nil {
		return ""
	}
	current := q.current()
	title := current.Question
	if current.Header != "" {
		title = current.Header + " · " + title
	}
	if n := len(q.req.Questions); n > 1 {
		title += fmt.Sprintf("  (%d/%d)", q.index+1, n)
	}
	wrap := max(m.width-4, 10)
	rows := []string{lipgloss.NewStyle().Bold(true).Width(wrap).Render(title)}
	for i, option := range current.Options {
		rows = append(rows, m.optionRow(q, i, option, wrap))
	}
	hint := singleHint
	if current.MultiSelect {
		hint = multiHint
	}
	return strings.Join(append(rows, m.theme.Muted.Width(wrap).Render(hint)), "\n")
}

// optionRow is one numbered option. A description that does not fit beside
// the label wraps under it, indented past the number.
func (m *Model) optionRow(q *question, i int, option askuser.Option, width int) string {
	marker := "  "
	if i == q.cursor {
		marker = "❯ "
	}
	if q.current().MultiSelect {
		box := "[ ] "
		if slices.Contains(q.answers[q.index].Picked, option.Label) {
			box = "[x] "
		}
		marker += box
	}
	row := lipgloss.NewStyle().Width(width).Render(fmt.Sprintf("%s%d. %s", marker, i+1, option.Label))
	if option.Description == "" {
		return row
	}
	if lipgloss.Height(row) == 1 && lipgloss.Width(strings.TrimRight(row, " "))+2+lipgloss.Width(option.Description) <= width {
		return strings.TrimRight(row, " ") + m.theme.Muted.Render("  "+option.Description)
	}
	indent := lipgloss.Width(marker) + 3
	desc := m.theme.Muted.PaddingLeft(indent).Width(width).Render(option.Description)
	return row + "\n" + desc
}

// questionRows is how many rows the pending question takes, counted from the
// rendered view, with the blank row under it.
func (m *Model) questionRows() int {
	view := m.questionView()
	if view == "" {
		return 0
	}
	return lipgloss.Height(view) + 1
}
