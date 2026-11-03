package runner

import (
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/nacelle"
)

func TestStampDatesTheRunPromptOnly(t *testing.T) {
	now := time.Date(2026, 11, 3, 14, 5, 0, 0, time.FixedZone("CET", 3600))
	messages := []nacelle.Message{
		{Role: nacelle.RoleUser, Parts: []nacelle.Part{nacelle.Text{Text: "older question"}}},
		{Role: nacelle.RoleAssistant, Parts: []nacelle.Part{nacelle.Text{Text: "answer"}}},
		{Role: nacelle.RoleUser, Parts: []nacelle.Part{nacelle.Text{Text: "what day is it?"}}},
		{Role: nacelle.RoleAssistant, Parts: []nacelle.Part{nacelle.ToolCall{ID: "1", Name: "x"}}},
		{Role: nacelle.RoleUser, Parts: []nacelle.Part{nacelle.ToolResult{ID: "1"}}},
	}
	stamp(messages, now)
	clock, ok := messages[2].Parts[0].(nacelle.Text)
	if !ok || !strings.Contains(clock.Text, "Tuesday 3 November 2026, 14:05 CET") {
		t.Fatalf("the run prompt is not stamped: %#v", messages[2].Parts)
	}
	if len(messages[0].Parts) != 1 || len(messages[4].Parts) != 1 {
		t.Fatal("a message other than the run prompt was stamped")
	}
}
