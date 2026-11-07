package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/core/nacelletest"
	"github.com/FacileStudio/nacelle"
)

func bridgedSet(t *testing.T, count int) []nacelle.Tool {
	t.Helper()
	set := make([]nacelle.Tool, 0, count)
	for i := range count {
		set = append(set, inputEcho(t, fmt.Sprintf("srv_tool_%02d", i)))
	}
	return set
}

func TestAGrownBridgedSetIsMountedAsACatalog(t *testing.T) {
	mounted := grown(nil, bridgedSet(t, mcpCatalogThreshold+1))
	var names []string
	for _, tool := range mounted {
		names = append(names, tool.Spec().Name)
	}
	if strings.Join(names, ",") != "search_tools,get_tool_details,call_tool" {
		t.Errorf("mounted %v, want the 3 catalog tools", names)
	}
}

func TestASmallBridgedSetIsMountedWhole(t *testing.T) {
	mounted := grown([]nacelle.Tool{}, bridgedSet(t, 1))
	if len(mounted) != 1 || mounted[0].Spec().Name != "srv_tool_00" {
		t.Errorf("mounted = %+v, want the bridged tool itself", mounted)
	}
}

func TestACatalogCallIsAskedByTheBridgedName(t *testing.T) {
	var seen []string
	approve := shapeCalls(nacelle.ApproveBool(func(_ context.Context, name string, input json.RawMessage) bool {
		seen = append(seen, name+" "+string(input))
		return name != "srv_tool_01"
	}), false)
	request := nacelle.Request{Tools: grown(nil, bridgedSet(t, mcpCatalogThreshold+1)), Approve: approve}
	ok, _, err := nacelletest.Call(t.Context(), request, nacelle.ToolCall{ID: "a", Name: "call_tool",
		Input: json.RawMessage(`{"name":"srv_tool_00","arguments":{"command":"hi"}}`)})
	if err != nil || ok.Err != nil || ok.Result != "hi" || ok.Name != "srv_tool_00" {
		t.Fatalf("allowed call: %+v %v", ok, err)
	}
	refused, _, err := nacelletest.Call(t.Context(), request, nacelle.ToolCall{ID: "b", Name: "call_tool",
		Input: json.RawMessage(`{"name":"srv_tool_01"}`)})
	if err != nil || refused.Err == nil {
		t.Fatalf("refused call: %+v %v", refused, err)
	}
	if strings.Join(seen, ",") != `srv_tool_00 {"command":"hi"},srv_tool_01 {}` {
		t.Fatalf("asked about %v", seen)
	}
}
