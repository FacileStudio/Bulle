package runner

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/FacileStudio/nacelle"
)

func TestDecodeMapsLegacyToolNames(t *testing.T) {
	raw := json.RawMessage(`{"role":"assistant","parts":[` +
		`{"type":"tool_call","id":"c1","name":"opus_task","input":{"id":7}},` +
		`{"type":"tool_result","id":"c1","name":"delete_background_job","result":"ok"}]}`)
	message, err := DecodeMessage(raw)
	if err != nil {
		t.Fatal(err)
	}
	call, _ := message.Parts[0].(nacelle.ToolCall)
	result, _ := message.Parts[1].(nacelle.ToolResult)
	if call.Name != "opus_get_task" || string(call.Input) != `{"task_id":7}` || result.Name != JobsDeleteTool {
		t.Fatalf("old names must replay under the new ones: %+v %+v", call, result)
	}
	wire, err := DecodeWire(raw)
	if err != nil || wire.Parts[0].Name != "opus_get_task" {
		t.Fatalf("the session view must show the new name too: %+v, %v", wire, err)
	}
}

func TestDecodeMapsPreV032NacelleToolNames(t *testing.T) {
	raw := json.RawMessage(`{"role":"assistant","parts":[` +
		`{"type":"tool_call","id":"c1","name":"run_command","input":{"command":"ls"}},` +
		`{"type":"tool_call","id":"c2","name":"search_content","input":{"pattern":"x"}},` +
		`{"type":"tool_result","id":"c3","name":"find_files","result":"a.go"}]}`)
	message, err := DecodeMessage(raw)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, part := range message.Parts {
		switch p := part.(type) {
		case nacelle.ToolCall:
			names = append(names, p.Name)
		case nacelle.ToolResult:
			names = append(names, p.Name)
		}
	}
	if strings.Join(names, ",") != "bash,grep,glob" {
		t.Fatalf("an old transcript must replay under the v0.32 names, got %v", names)
	}
}
