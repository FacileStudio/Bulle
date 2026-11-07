package approval

import (
	"context"
	"encoding/json"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/FacileStudio/Bulle/apps/cli/internal/tasks"
)

func TestThePlanToolIsNeverPutInFrontOfAHuman(t *testing.T) {
	gate := New()
	gate.send = func(msg tea.Msg) {
		t.Errorf("the plan tool was put in front of a human: %#v", msg)
		if req, ok := msg.(Request); ok {
			req.Decision <- Denied
		}
	}

	input := json.RawMessage(`{"todos":[{"content":"read the file","status":"in_progress"}]}`)
	if !gate.allows(context.Background(), tasks.NewTool().Spec().Name, input) {
		t.Fatal("the plan tool was refused")
	}
}

func TestThePlanToolIsStillRefusedWhenItsInputIsAmbiguous(t *testing.T) {
	gate := New()
	gate.send = func(tea.Msg) {
		t.Fatal("an unrenderable call was put in front of a human anyway")
	}

	input := json.RawMessage(`{"todos":[],"todos":[{"content":"x","status":"pending"}]}`)
	if gate.allows(context.Background(), tasks.NewTool().Spec().Name, input) {
		t.Error("a plan whose input has two values for one key was approved")
	}
}
