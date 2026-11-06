package suite

import "testing"

type scopeCase struct {
	name  string
	input string
	want  string
}

var scopeDefaults = Defaults{CourrierAccount: 5, AgendaCalendar: 7, OpusWorkspace: "ws1"}

var scopeCases = []scopeCase{
	{"opus_create_task", `{"project":"p9","title":"T"}`, "workspace:ws1/project:p9"},
	{"opus_create_task", `{"title":"T"}`, "workspace:ws1"},
	{"opus_create_task", `{"project":"p9","workspace":"wsX"}`, "workspace:wsX/project:p9"},
	{"opus_update_task", `{"id":1}`, "task:1"},
	{"opus_delete_task", `{"id":2}`, "task:2"},
	{"opus_create_project", `{"workspace":"wsX","name":"N"}`, "workspace:wsX"},
	{"opus_update_project", `{"id":3}`, "project:3"},
	{"opus_delete_project", `{"id":4}`, "project:4"},
	{"agenda_create_event", `{"title":"x"}`, "calendar:7"},
	{"agenda_create_event", `{"calendar":9,"title":"x"}`, "calendar:9"},
	{"agenda_update_event", `{"id":11}`, "event:11"},
	{"agenda_delete_event", `{"id":12}`, "event:12"},
	{"sablier_start", `{"project":2,"task":3}`, "project:2/task:3"},
	{"sablier_stop", `{}`, "timer"},
	{"courrier_send", `{"body":"hi"}`, "account:5"},
	{"courrier_save_draft", `{"body":"hi"}`, "account:5"},
	{"courrier_save_draft", `{"id":8,"body":"hi"}`, "account:5/draft:8"},
	{"courrier_delete_draft", `{"id":8}`, "account:5/draft:8"},
	{"courrier_update", `{"ids":[1,2],"action":"mark_read"}`, "account:5/action:mark_read"},
	{"courrier_update", `{"ids":[1],"action":"delete"}`, ""},
	{"courrier_move", `{"ids":[1],"folder":"inbox"}`, "account:5/move:inbox"},
	{"courrier_move", `{"ids":[1],"folder_id":44}`, "account:5/move:44"},
	{"files_save", `{"content":"hi","name":"n.txt"}`, "folder:0"},
	{"glouton_create_company", `{"name":"Acme"}`, "companies"},
	{"glouton_update_company", `{"id":"c1"}`, "company:c1"},
	{"glouton_create_person", `{"firstName":"Ann"}`, "people"},
	{"glouton_create_person", `{"firstName":"Ann","companyId":"c1"}`, "company:c1"},
	{"glouton_update_person", `{"id":"p1"}`, "person:p1"},
}

func TestScopeForTargets(t *testing.T) {
	for _, tc := range scopeCases {
		scope, summary, err := ScopeFor(tc.name, []byte(tc.input), scopeDefaults)
		if err != nil {
			t.Errorf("ScopeFor(%s, %s): %v", tc.name, tc.input, err)
			continue
		}
		if scope != tc.want {
			t.Errorf("ScopeFor(%s, %s) = %q, want %q", tc.name, tc.input, scope, tc.want)
		}
		if summary["resource"] == "" || summary["action"] == "" {
			t.Errorf("ScopeFor(%s, %s) summary lacks resource or action: %v", tc.name, tc.input, summary)
		}
	}
}

func TestScopeForCoversEveryWriteTool(t *testing.T) {
	covered := map[string]bool{}
	for _, tc := range scopeCases {
		if tc.want != "" {
			covered[tc.name] = true
		}
	}
	for _, name := range WriteTools() {
		if !covered[name] {
			t.Errorf("%s has no documented scope, so it would be asked every time", name)
		}
	}
}

func TestScopeForNamespacesResources(t *testing.T) {
	pairs := []scopeCase{
		{"opus_update_task", `{"id":7}`, "task:7"},
		{"opus_update_project", `{"id":7}`, "project:7"},
		{"agenda_update_event", `{"id":7}`, "event:7"},
		{"glouton_update_company", `{"id":"7"}`, "company:7"},
		{"glouton_update_person", `{"id":"7"}`, "person:7"},
	}
	seen := map[string]string{}
	for _, tc := range pairs {
		scope, _, err := ScopeFor(tc.name, []byte(tc.input), scopeDefaults)
		if err != nil {
			t.Fatalf("ScopeFor(%s): %v", tc.name, err)
		}
		if scope != tc.want {
			t.Errorf("ScopeFor(%s) = %q, want %q", tc.name, scope, tc.want)
		}
		if other, clash := seen[scope]; clash {
			t.Errorf("scope %q is shared by %s and %s", scope, other, tc.name)
		}
		seen[scope] = tc.name
	}
}

func TestScopeForFailures(t *testing.T) {
	unreadable := []scopeCase{
		{"opus_update_task", `not json`, ""},
		{"opus_update_task", `{"id":"abc"}`, ""},
		{"courrier_update", `{"ids":"one","action":"delete"}`, ""},
		{"unknown_tool", `{}`, ""},
	}
	for _, tc := range unreadable {
		if _, _, err := ScopeFor(tc.name, []byte(tc.input), scopeDefaults); err == nil {
			t.Errorf("ScopeFor(%s, %s) accepted an unreadable call", tc.name, tc.input)
		}
	}
	untargeted := []scopeCase{
		{"opus_update_task", `{"id":0}`, ""},
		{"opus_create_task", `{"project":"p9"}`, ""},
		{"opus_create_project", `{"name":"N"}`, ""},
		{"agenda_create_event", `{}`, ""},
		{"courrier_update", `{"ids":[1],"action":"delete"}`, ""},
		{"courrier_move", `{"ids":[1]}`, ""},
		{"sablier_start", `{"project":0,"task":3}`, ""},
	}
	for _, tc := range untargeted {
		scope, summary, err := ScopeFor(tc.name, []byte(tc.input), Defaults{})
		if err != nil {
			t.Errorf("ScopeFor(%s, %s): %v", tc.name, tc.input, err)
			continue
		}
		if scope != tc.want {
			t.Errorf("ScopeFor(%s, %s) = %q, want an empty scope", tc.name, tc.input, scope)
		}
		if summary["resource"] == "" {
			t.Errorf("ScopeFor(%s, %s) lost its record: %v", tc.name, tc.input, summary)
		}
	}
}
