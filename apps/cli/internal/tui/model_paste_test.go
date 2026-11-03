package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// Pasted content arrives as one message, so a multi-line paste must keep its
// newlines rather than being collapsed onto a single line.
func TestHandlePastePreservesMultilineContent(t *testing.T) {
	m := bareBanner()
	m.prompt.InsertString("initial")
	m.promptRoute(tea.PasteMsg{Content: "line1\nline2\nline3"})
	if got := m.prompt.Value(); got != "initialline1\nline2\nline3" {
		t.Errorf("handlePaste = %q, want %q", got, "initialline1\nline2\nline3")
	}
}

// Windows line endings are normalized to Unix style before being inserted.
func TestHandlePasteNormalizesWindowsLineEndings(t *testing.T) {
	m := bareBanner()
	m.promptRoute(tea.PasteMsg{Content: "line1\r\nline2"})
	if got := m.prompt.Value(); got != "line1\nline2" {
		t.Errorf("handlePaste = %q, want %q", got, "line1\nline2")
	}
}

// Old Mac line endings are normalized to Unix style before being inserted.
func TestHandlePasteNormalizesMacLineEndings(t *testing.T) {
	m := bareBanner()
	m.promptRoute(tea.PasteMsg{Content: "line1\rline2"})
	if got := m.prompt.Value(); got != "line1\nline2" {
		t.Errorf("handlePaste = %q, want %q", got, "line1\nline2")
	}
}

func longPaste(lines int) string {
	return strings.TrimSuffix(strings.Repeat("row\n", lines), "\n")
}

// A long paste shows as a placeholder and goes out as the full text.
func TestLongPasteCollapsesAndExpandsOnSubmit(t *testing.T) {
	m := bareBanner()
	m.prompt.InsertString("look: ")
	m.promptRoute(tea.PasteMsg{Content: longPaste(133)})
	if got := m.prompt.Value(); got != "look: [Pasted text #1 +133 lines]" {
		t.Fatalf("prompt = %q, want the placeholder", got)
	}
	if got := m.takePrompt(); got != "look: "+longPaste(133) {
		t.Fatalf("taken = %q, want the expanded paste", got)
	}
	if len(m.pasted.texts) != 0 {
		t.Fatal("pastes survived the submit")
	}
}

// A paste the terminal delivered in two chunks merges into one placeholder.
func TestSplitPasteMergesIntoOnePlaceholder(t *testing.T) {
	m := bareBanner()
	m.promptRoute(tea.PasteMsg{Content: longPaste(12) + "\n"})
	m.promptRoute(tea.PasteMsg{Content: longPaste(5)})
	if got := m.prompt.Value(); got != "[Pasted text #1 +17 lines]" {
		t.Fatalf("prompt = %q, want one merged placeholder", got)
	}
	if got := m.takePrompt(); got != longPaste(12)+"\n"+longPaste(5) {
		t.Fatalf("taken = %q, want both chunks", got)
	}
}

// Typing between two pastes keeps them apart, numbered in order.
func TestSeparatePastesKeepTheirOwnNumbers(t *testing.T) {
	m := bareBanner()
	m.promptRoute(tea.PasteMsg{Content: longPaste(11)})
	m.prompt.InsertString(" vs ")
	m.promptRoute(tea.PasteMsg{Content: longPaste(20)})
	if got := m.prompt.Value(); got != "[Pasted text #1 +11 lines] vs [Pasted text #2 +20 lines]" {
		t.Fatalf("prompt = %q", got)
	}
}

func TestEchoKeepsALongQuestionShort(t *testing.T) {
	if got := echo(longPaste(40)); got != "row\nrow\nrow\n… +37 lines" {
		t.Fatalf("echo = %q", got)
	}
	if got := echo("short\nquestion"); got != "short\nquestion" {
		t.Fatalf("echo = %q, want a short question untouched", got)
	}
}
