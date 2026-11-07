package tui

import (
	"testing"

	"github.com/FacileStudio/nacelle"
)

// /copy hands the newest answer with text to the clipboard, skipping a
// tool-only assistant turn after it.
func TestCopyTakesTheLastAnswer(t *testing.T) {
	m := sized()
	m.conversation = []nacelle.Message{
		nacelle.UserText("q"),
		nacelle.AssistantText("the answer"),
		{Role: nacelle.RoleAssistant},
	}
	if got := newestAnswer(m.conversation); got != "the answer" {
		t.Fatalf("newestAnswer = %q", got)
	}
	cmd, ok := m.parseCommand("/copy")
	if !ok || cmd(m) == nil {
		t.Error("/copy returned no clipboard command")
	}
	if newestAnswer(nil) != "" {
		t.Error("an empty conversation has an answer")
	}
}
