package tui

import (
	"strings"
	"testing"
)

// The inline cursor sits on the row the typed text is drawn on, whatever the
// live region above it holds.
func TestInlineCursorLandsOnThePromptRow(t *testing.T) {
	for _, answer := range []string{"", "a streamed answer"} {
		m := sized()
		m.run.answer.WriteString(answer)
		m.prompt.SetValue("zqxj")
		view := m.assembleInline()
		if view.Cursor == nil {
			t.Fatal("no cursor")
		}
		rows := strings.Split(view.Content, "\n")
		if y := view.Cursor.Y; y >= len(rows) || !strings.Contains(visible(rows[y]), "zqxj") {
			t.Errorf("answer %q: cursor row %d is not the prompt row", answer, y)
		}
	}
}
