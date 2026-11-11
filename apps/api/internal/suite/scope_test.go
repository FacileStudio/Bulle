package suite

import "testing"

type scopeCase struct {
	name  string
	input string
	want  string
}

var scopeDefaults = Defaults{CourrierAccount: 5}

var scopeCases = []scopeCase{
	{"glouton_create_company", `{"name":"Acme"}`, "companies"},
	{"glouton_update_company", `{"company_id":"c1"}`, "company:c1"},
	{"glouton_create_person", `{"first_name":"Ann"}`, "people"},
	{"glouton_create_person", `{"first_name":"Ann","company_id":"c1"}`, "company:c1"},
	{"glouton_update_person", `{"person_id":"p1"}`, "person:p1"},
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
		{"glouton_update_company", `not json`, ""},
		{"unknown_tool", `{}`, ""},
	}
	for _, tc := range unreadable {
		if _, _, err := ScopeFor(tc.name, []byte(tc.input), scopeDefaults); err == nil {
			t.Errorf("ScopeFor(%s, %s) accepted an unreadable call", tc.name, tc.input)
		}
	}
	untargeted := []scopeCase{
		{"glouton_update_company", `{"company_id":""}`, ""},
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
