package agent

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

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

func TestShapeCallsAsksCallToolByTheBridgedName(t *testing.T) {
	var seen []string
	approve := shapeCalls(nacelle.ApproveBool(func(_ context.Context, name string, _ json.RawMessage) bool {
		seen = append(seen, name)
		return name != "srv_wipe"
	}), false)
	ok := approve(context.Background(), nacelle.Call{Name: callTool, Input: json.RawMessage(`{"name":"srv_echo","arguments":{"hi":true}}`)})
	refused := approve(context.Background(), nacelle.Call{Name: callTool, Input: json.RawMessage(`{"name":"srv_wipe"}`)})
	unnamed := approve(context.Background(), nacelle.Call{Name: callTool, Input: json.RawMessage(`{"arguments":{}}`)})
	if ok.Deny != "" || ok.Input != nil || ok.Name != "" || refused.Deny == "" || unnamed.Deny != "" {
		t.Fatalf("decisions %+v %+v %+v", ok, refused, unnamed)
	}
	if strings.Join(seen, ",") != "srv_echo,srv_wipe,call_tool" {
		t.Fatalf("asked about %v", seen)
	}
}

func TestShapeCallsCoAuthorsACommitThroughTheSink(t *testing.T) {
	for _, on := range []bool{true, false} {
		tool := inputEcho(t, shellTool)
		sink := &nacelle.ToolSink{Approve: shapeCalls(nil, on), Tools: nacelle.ToolsByName([]nacelle.Tool{tool})}
		out, err := nacelle.RunTool(context.Background(), tool, nacelle.Invocation{ID: "c1", Name: shellTool}, json.RawMessage(`{"command":"git commit -m x"}`), sink)
		if err != nil {
			t.Fatalf("run: %v", err)
		}
		if strings.Contains(out, coAuthorTrailer) != on {
			t.Fatalf("coauthor %v ran %q", on, out)
		}
	}
}
