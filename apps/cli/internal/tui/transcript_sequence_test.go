package tui

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/FacileStudio/Bulle/apps/cli/internal/ask"
	"github.com/FacileStudio/Bulle/core/askuser"
	"github.com/FacileStudio/nacelle"
)

// saidIndex is where needle first appears in what the model has said so far.
func saidIndex(m *Model, needle string) int {
	return strings.Index(visible(strings.Join(m.unprinted, "\n")), needle)
}

// The reader's answers print under the ask_user tool line, not above it.
func TestAskReplyPrintsAfterTheToolLine(t *testing.T) {
	m := sized()
	call := &nacelle.ToolEvent{ID: "q1", Name: askuser.Name, Input: `{"questions":[]}`}
	m.absorb(nacelle.Event{Kind: nacelle.KindToolCall, Tool: call})
	q := []askuser.Question{{Question: "Which day?", Options: []askuser.Option{{Label: "Monday"}, {Label: "Tuesday"}}}}
	m.Update(ask.Request{Questions: q, Answers: make(chan []askuser.Answer, 1)})
	m.Update(tea.KeyPressMsg{Code: '1', Text: "1"})
	m.absorb(nacelle.Event{Kind: nacelle.KindToolResult, Tool: call})

	tool, reply := saidIndex(m, askuser.Name), saidIndex(m, "Which day? Monday")
	if tool < 0 || reply < 0 || reply < tool {
		t.Errorf("tool line at %d, reply at %d; want the reply after the line", tool, reply)
	}
}

// A failure held for collapsing prints before the text that follows it.
func TestDeferredFailurePrintsBeforeLaterText(t *testing.T) {
	m := sized()
	call := &nacelle.ToolEvent{ID: "f1", Name: "read_file", Input: `{"path":"gone.txt"}`}
	m.absorb(nacelle.Event{Kind: nacelle.KindToolCall, Tool: call})
	failed := *call
	failed.Err = errors.New("no such file")
	m.absorb(nacelle.Event{Kind: nacelle.KindToolResult, Tool: &failed})
	m.absorb(nacelle.Event{Kind: nacelle.KindText, Text: "Trying another path.\n\nnext"})

	failure, text := saidIndex(m, "no such file"), saidIndex(m, "Trying another path")
	if failure < 0 || text < 0 || failure > text {
		t.Errorf("failure at %d, text at %d; want the failure first", failure, text)
	}
}
