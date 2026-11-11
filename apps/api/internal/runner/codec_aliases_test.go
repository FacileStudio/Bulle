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
	if call.Name != "opus_read_task" || string(call.Input) != `{"task_id":7}` || result.Name != JobsDeleteTool {
		t.Fatalf("old names must replay under the new ones: %+v %+v", call, result)
	}
	wire, err := DecodeWire(raw)
	if err != nil || wire.Parts[0].Name != "opus_read_task" {
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

func TestDecodeFinishesACallStoredBeforeTheFlag(t *testing.T) {
	legacy, err := DecodeMessage(json.RawMessage(`{"role":"assistant","parts":[{"type":"tool_call","id":"c1","name":"run_command","input":{"command":"ls"}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if call, _ := legacy.Parts[0].(nacelle.ToolCall); !call.Finished {
		t.Fatalf("a call stored before the finished flag must replay as finished: %+v", call)
	}
	cut, err := DecodeMessage(json.RawMessage(`{"role":"assistant","parts":[{"type":"tool_call","id":"c2","name":"bash","input":{},"finished":false}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if call, _ := cut.Parts[0].(nacelle.ToolCall); call.Finished {
		t.Fatalf("a call stored as unfinished must stay unfinished: %+v", call)
	}
}

func TestDecodeShowsBuiltInCourrierCallsUnderTheMCPNames(t *testing.T) {
	raw := json.RawMessage(`{"role":"assistant","parts":[` +
		`{"type":"tool_call","id":"c1","name":"courrier_list_inbox","input":{"account":5,"unread":true}},` +
		`{"type":"tool_call","id":"c2","name":"courrier_get_email","input":{"email_id":9,"account":5}},` +
		`{"type":"tool_call","id":"c3","name":"courrier_mark","input":{"email_ids":[9],"action":"mark_read"}},` +
		`{"type":"tool_result","id":"c2","name":"courrier_get_email","result":"hello"}]}`)
	message, err := DecodeMessage(raw)
	if err != nil {
		t.Fatal(err)
	}
	want := []struct{ name, input string }{
		{"courrier_search_mail", `{"account_id":5,"unread_only":true}`},
		{"courrier_read_thread", `{"account_id":5,"message_id":9}`},
		{"courrier_organize", `{"action":"mark_read","message_ids":[9]}`},
	}
	for i, w := range want {
		call, _ := message.Parts[i].(nacelle.ToolCall)
		if call.Name != w.name || string(call.Input) != w.input {
			t.Errorf("part %d = %s %s, want %s %s", i, call.Name, call.Input, w.name, w.input)
		}
	}
	if result, _ := message.Parts[3].(nacelle.ToolResult); result.Name != "courrier_read_thread" || !untrusted(result.Name) {
		t.Fatalf("an old mail read must replay as a source of untrusted text: %+v", result)
	}
	wire, err := DecodeWire(raw)
	if err != nil || wire.Parts[0].Name != "courrier_search_mail" {
		t.Fatalf("the session view must show the MCP name: %+v, %v", wire, err)
	}
}

func TestDecodeShowsBuiltInAgendaCallsUnderTheMCPNames(t *testing.T) {
	raw := json.RawMessage(`{"role":"assistant","parts":[` +
		`{"type":"tool_call","id":"c1","name":"agenda_get_event","input":{"id":4}},` +
		`{"type":"tool_call","id":"c2","name":"agenda_update_event","input":{"event_id":4,"start_at":"2026-11-12T14:00:00+01:00","calendar":3}},` +
		`{"type":"tool_call","id":"c3","name":"agenda_events","input":{"from":"today"}},` +
		`{"type":"tool_result","id":"c1","name":"agenda_get_event","result":"standup"}]}`)
	message, err := DecodeMessage(raw)
	if err != nil {
		t.Fatal(err)
	}
	want := []struct{ name, input string }{
		{"agenda_read_event", `{"event_id":4}`},
		{"agenda_update_event", `{"calendar_id":3,"event_id":4,"start":"2026-11-12T14:00:00+01:00"}`},
		{"agenda_list_events", `{"from":"today"}`},
	}
	for i, w := range want {
		call, _ := message.Parts[i].(nacelle.ToolCall)
		if call.Name != w.name || string(call.Input) != w.input {
			t.Errorf("part %d = %s %s, want %s %s", i, call.Name, call.Input, w.name, w.input)
		}
	}
	if result, _ := message.Parts[3].(nacelle.ToolResult); result.Name != "agenda_read_event" || !untrusted(result.Name) {
		t.Fatalf("an old event read must replay as a source of untrusted text: %+v", result)
	}
}

func TestDecodeShowsBuiltInWebCallsUnderTheSonarNames(t *testing.T) {
	raw := json.RawMessage(`{"role":"assistant","parts":[` +
		`{"type":"tool_call","id":"c1","name":"web_search","input":{"query":"go 1.27","count":3}},` +
		`{"type":"tool_call","id":"c2","name":"web_fetch","input":{"url":"https://go.dev/doc/go1.27"}},` +
		`{"type":"tool_result","id":"c2","name":"web_fetch","result":"Go 1.27 release notes"}]}`)
	message, err := DecodeMessage(raw)
	if err != nil {
		t.Fatal(err)
	}
	want := []struct{ name, input string }{
		{"sonar_search", `{"query":"go 1.27","count":3}`},
		{"sonar_fetch", `{"url":"https://go.dev/doc/go1.27"}`},
	}
	for i, w := range want {
		call, _ := message.Parts[i].(nacelle.ToolCall)
		if call.Name != w.name || string(call.Input) != w.input {
			t.Errorf("part %d = %s %s, want %s %s", i, call.Name, call.Input, w.name, w.input)
		}
	}
	if result, _ := message.Parts[2].(nacelle.ToolResult); result.Name != "sonar_fetch" || !untrusted(result.Name) {
		t.Fatalf("an old page read must replay as a source of untrusted text: %+v", result)
	}
}

func TestDecodeShowsBuiltInNuageCallsUnderTheMCPNames(t *testing.T) {
	raw := json.RawMessage(`{"role":"assistant","parts":[` +
		`{"type":"tool_call","id":"c1","name":"files_read","input":{"id":12}},` +
		`{"type":"tool_call","id":"c2","name":"nuage_list","input":{"folder":3}},` +
		`{"type":"tool_call","id":"c3","name":"nuage_save","input":{"folder":3,"name":"n.md"}},` +
		`{"type":"tool_result","id":"c1","name":"nuage_read","result":"notes"}]}`)
	message, err := DecodeMessage(raw)
	if err != nil {
		t.Fatal(err)
	}
	want := []struct{ name, input string }{
		{"nuage_read_file", `{"file_id":12}`},
		{"nuage_list_folder", `{"folder_id":3}`},
		{"nuage_save_file", `{"folder_id":3,"name":"n.md"}`},
	}
	for i, w := range want {
		call, _ := message.Parts[i].(nacelle.ToolCall)
		if call.Name != w.name || string(call.Input) != w.input {
			t.Errorf("part %d = %s %s, want %s %s", i, call.Name, call.Input, w.name, w.input)
		}
	}
	if result, _ := message.Parts[3].(nacelle.ToolResult); result.Name != "nuage_read_file" || !untrusted(result.Name) {
		t.Fatalf("an old file read must replay as a source of untrusted text: %+v", result)
	}
}

func TestDecodeShowsBuiltInOpusCallsUnderTheMCPNames(t *testing.T) {
	raw := json.RawMessage(`{"role":"assistant","parts":[` +
		`{"type":"tool_call","id":"c1","name":"opus_list_tasks","input":{"status":"to-do"}},` +
		`{"type":"tool_call","id":"c2","name":"opus_update_task","input":{"id":7,"due_date":"2026-11-30"}},` +
		`{"type":"tool_call","id":"c3","name":"opus_get_project","input":{"project_id":5}},` +
		`{"type":"tool_result","id":"c1","name":"opus_list_tasks","result":"tasks"}]}`)
	message, err := DecodeMessage(raw)
	if err != nil {
		t.Fatal(err)
	}
	want := []struct{ name, input string }{
		{"opus_search_tasks", `{"status":"to-do"}`},
		{"opus_update_task", `{"due":"2026-11-30","task_id":7}`},
		{"opus_list_projects", `{"project_id":5}`},
	}
	for i, w := range want {
		call, _ := message.Parts[i].(nacelle.ToolCall)
		if call.Name != w.name || string(call.Input) != w.input {
			t.Errorf("part %d = %s %s, want %s %s", i, call.Name, call.Input, w.name, w.input)
		}
	}
	if result, _ := message.Parts[3].(nacelle.ToolResult); result.Name != "opus_search_tasks" || !untrusted(result.Name) {
		t.Fatalf("an old task list must replay as a source of untrusted text: %+v", result)
	}
}
