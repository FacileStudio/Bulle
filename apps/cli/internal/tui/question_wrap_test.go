package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/FacileStudio/Bulle/apps/cli/internal/ask"
	"github.com/FacileStudio/Bulle/core/askuser"
)

// At a narrow width the title and a long description wrap inside width-4, and
// questionRows counts the wrapped rows, not one per option.
func TestQuestionWrapsAndCountsRenderedRows(t *testing.T) {
	m := sized()
	m.resize(tea.WindowSizeMsg{Width: 40, Height: 30})
	long := strings.Repeat("a long description ", 6)
	q := []askuser.Question{{Question: "Which of these deployment strategies should the release use?", Options: []askuser.Option{{Label: "Blue", Description: long}, {Label: "Green"}}}}
	m.Update(ask.Request{Questions: q, Answers: make(chan []askuser.Answer, 1)})

	view := m.questionView()
	for _, row := range strings.Split(view, "\n") {
		if w := lipgloss.Width(row); w > 36 {
			t.Errorf("row %q is %d wide, want <= 36", visible(row), w)
		}
	}
	if !strings.Contains(strings.Join(strings.Fields(visible(view)), " "), strings.TrimSpace(long)) {
		t.Errorf("description lost in the wrap:\n%s", visible(view))
	}
	if got, want := m.questionRows(), lipgloss.Height(view)+1; got != want || got <= len(q[0].Options)+3 {
		t.Errorf("questionRows = %d, want the %d rendered rows", got, want)
	}
}
