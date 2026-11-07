package runner

import (
	"encoding/json"
	"testing"

	"github.com/FacileStudio/nacelle"
)

func TestACappedRunIsStoredWithoutItsUnansweredCalls(t *testing.T) {
	capped := []nacelle.Message{
		nacelle.UserText("tidy my inbox"),
		{Role: nacelle.RoleAssistant, Parts: []nacelle.Part{
			nacelle.ToolCall{ID: "c1", Name: "courrier_list_inbox", Input: json.RawMessage(`{}`), Finished: true},
			nacelle.Finish{Stop: nacelle.StopIterations},
		}},
	}
	stored := (&Runner{}).transcript(capped, outcome{status: "finished", stop: string(nacelle.StopIterations)})
	for _, message := range stored {
		for _, part := range message.Parts {
			if _, ok := part.(nacelle.ToolCall); ok {
				t.Fatalf("an unanswered call was stored: %+v", stored)
			}
		}
	}
	if len(stored) != 2 || stored[1].Role != nacelle.RoleAssistant {
		t.Fatalf("want the question and a note that the run stopped, got %+v", stored)
	}
}
