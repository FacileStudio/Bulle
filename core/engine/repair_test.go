package engine

import (
	"testing"

	"github.com/FacileStudio/nacelle"
)

func TestRepairDropsACallNoResultAnswers(t *testing.T) {
	capped := []nacelle.Message{
		nacelle.UserText("question"),
		{Role: nacelle.RoleAssistant, Parts: []nacelle.Part{
			nacelle.Text{Text: "let me look"},
			nacelle.ToolCall{ID: "orphan", Name: "read", Finished: true},
			nacelle.Finish{Stop: nacelle.StopIterations},
		}},
	}
	got := Repair(capped)
	if roles(got) != "user,assistant" {
		t.Fatalf("roles = %s", roles(got))
	}
	for _, part := range got[1].Parts {
		if _, ok := part.(nacelle.ToolCall); ok {
			t.Fatalf("the unanswered call was kept: %+v", got[1])
		}
	}
}

func TestRepairDropsATurnLeftWithNothingToSay(t *testing.T) {
	got := Repair([]nacelle.Message{
		nacelle.UserText("first"),
		{Role: nacelle.RoleAssistant, Parts: []nacelle.Part{
			nacelle.Reasoning{Text: "hm"},
			nacelle.ToolCall{ID: "orphan", Name: "read", Finished: true},
			nacelle.Finish{Stop: nacelle.StopIterations},
		}},
		nacelle.UserText("second"),
	})
	if len(got) != 1 || len(got[0].Parts) != 2 {
		t.Fatalf("want the two user turns merged into one message, got %+v", got)
	}
}

func TestRepairKeepsAnAnsweredCall(t *testing.T) {
	got := Repair([]nacelle.Message{
		nacelle.UserText("read file"),
		{Role: nacelle.RoleAssistant, Parts: []nacelle.Part{nacelle.ToolCall{ID: "c1", Name: "read", Finished: true}}},
		{Role: nacelle.RoleUser, Parts: []nacelle.Part{nacelle.ToolResult{ID: "c1", Name: "read", Result: "data"}}},
		nacelle.AssistantText("done"),
	})
	if roles(got) != "user,assistant,user,assistant" {
		t.Fatalf("roles = %s", roles(got))
	}
}
