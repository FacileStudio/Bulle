package runner

import (
	"encoding/json"
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
