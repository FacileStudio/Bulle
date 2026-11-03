package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestAltEnterInsertsNewlineAfterPaste(t *testing.T) {
	m := bareBanner()
	m.promptRoute(tea.PasteMsg{Content: "pasted content"})
	if got := m.prompt.Value(); got != "pasted content" {
		t.Fatalf("paste = %q, want %q", got, "pasted content")
	}

	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter, Mod: tea.ModAlt})
	if got := m.prompt.Value(); got != "pasted content\n" {
		t.Errorf("alt+enter after paste = %q, want %q", got, "pasted content\n")
	}
}

// More lines than the prompt's viewport (a recalled question, a short paste
// in a small prompt) leave the textarea at its content cap; bubbles then
// silently drops every newline insertion, so alt+enter must keep working.
func TestAltEnterAfterLongPaste(t *testing.T) {
	m := bareBanner()
	lines := make([]string, promptRows+5)
	for i := range lines {
		lines[i] = fmt.Sprintf("line %d", i)
	}
	m.insertLines(strings.Join(lines, "\n"))
	if got := strings.Count(m.prompt.Value(), "\n") + 1; got != len(lines) {
		t.Fatalf("insert kept %d of %d lines", got, len(lines))
	}

	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter, Mod: tea.ModAlt})
	if got := strings.Count(m.prompt.Value(), "\n") + 1; got != len(lines)+1 {
		t.Errorf("alt+enter after long paste kept %d lines, want %d", got, len(lines)+1)
	}
}

func TestShiftEnterInsertsNewline(t *testing.T) {
	m := bareBanner()
	m.prompt.SetValue("first line")

	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter, Mod: tea.ModShift})
	if got := m.prompt.Value(); got != "first line\n" {
		t.Errorf("shift+enter = %q, want %q", got, "first line\n")
	}
	m.prompt.InsertString("second")
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter, Mod: tea.ModShift})
	if got := m.prompt.Value(); got != "first line\nsecond\n" {
		t.Errorf("shift+enter = %q, want %q", got, "first line\nsecond\n")
	}
}
