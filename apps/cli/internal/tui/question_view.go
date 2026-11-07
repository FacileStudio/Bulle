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
// one numbered row per option, and the keys that answer it.
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
	rows := []string{lipgloss.NewStyle().Bold(true).Render(title)}
	for i, option := range current.Options {
		rows = append(rows, m.optionRow(q, i, option))
	}
	hint := singleHint
	if current.MultiSelect {
		hint = multiHint
	}
	return strings.Join(append(rows, m.theme.Muted.Render(hint)), "\n")
}

func (m *Model) optionRow(q *question, i int, option askuser.Option) string {
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
	row := fmt.Sprintf("%s%d. %s", marker, i+1, option.Label)
	if option.Description != "" {
		row += m.theme.Muted.Render("  " + option.Description)
	}
	return row
}

// questionRows is how many rows the pending question takes, with the blank
// row under it.
func (m *Model) questionRows() int {
	if m.run.question == nil {
		return 0
	}
	return len(m.run.question.current().Options) + 3
}
