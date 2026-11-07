package agent

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/core/nacelletest"
	"github.com/FacileStudio/nacelle"
)

type echoInput struct {
	Command string `json:"command"`
}

func inputEcho(t *testing.T, name string) nacelle.Tool {
	t.Helper()
	tool, err := nacelle.NewTool(name, name, func(_ context.Context, in echoInput) (string, error) {
		return in.Command, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return tool
}

func TestShapeCallsCoAuthorsACommitThroughTheLoop(t *testing.T) {
	for _, on := range []bool{true, false} {
		tool := inputEcho(t, shellTool)
		request := nacelle.Request{Tools: []nacelle.Tool{tool}, Approve: shapeCalls(nil, on)}
		event, _, err := nacelletest.Call(context.Background(), request,
			nacelle.ToolCall{ID: "c1", Name: shellTool, Input: json.RawMessage(`{"command":"git commit -m x"}`)})
		if err != nil || event.Err != nil {
			t.Fatalf("run: %v %v", err, event)
		}
		out := event.Result
		if strings.Contains(out, coAuthorTrailer) != on {
			t.Fatalf("coauthor %v ran %q", on, out)
		}
	}
}
