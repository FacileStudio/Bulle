package suite

import (
	"encoding/json"
	"strings"
	"testing"
)

func registeredNames(t *testing.T) map[string]bool {
	t.Helper()
	w := newWorld(t)
	tools, _, err := Tools(NewGateway(w.tokens(), w.cfg, nil), 1, Defaults{}, Runtime{})
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, tool := range tools {
		names[tool.Spec().Name] = true
		schema, _ := json.Marshal(tool.Spec().Schema)
		for old := range legacyArgs[tool.Spec().Name] {
			if strings.Contains(string(schema), `"`+old+`":{`) {
				t.Errorf("%s still advertises the old argument %s", tool.Spec().Name, old)
			}
		}
	}
	return names
}

func TestLegacyNamesPointAtRegisteredTools(t *testing.T) {
	names := registeredNames(t)
	for old, current := range legacyToolNames {
		if names[old] {
			t.Errorf("%s is still registered under its old name", old)
		}
		if !names[current] && !strings.HasPrefix(current, "jobs_") {
			t.Errorf("%s maps to %s, which is not registered", old, current)
		}
	}
	for name := range legacyArgs {
		if !names[name] && !strings.HasPrefix(name, "jobs_") {
			t.Errorf("argument aliases for %s, which is not registered", name)
		}
	}
}

func TestCanonicalInputRenamesOldArguments(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"opus_task", `{"id":5}`, `{"task_id":5}`},
		{"opus_update_task", `{"id":5,"task_id":6}`, `{"id":5,"task_id":6}`},
		{"glouton_create_company", `{"name":"A","people":[{"firstName":"Ann"}]}`, `{"name":"A","people":[{"first_name":"Ann"}]}`},
		{"courrier_move", `{"ids":[1]}`, `{"email_ids":[1]}`},
		{"courrier_search", `{"query":"x"}`, `{"query":"x"}`},
		{"opus_get_task", `[1]`, `[1]`},
	}
	for _, tc := range cases {
		if got := string(CanonicalInput(tc.name, json.RawMessage(tc.in))); got != tc.want {
			t.Errorf("CanonicalInput(%s, %s) = %s, want %s", tc.name, tc.in, got, tc.want)
		}
	}
	if CanonicalTool("files_save") != "nuage_save" || CanonicalTool("courrier_send") != "courrier_send" {
		t.Fatal("CanonicalTool must map old names and keep current ones")
	}
}

func TestOldArgumentsStillRun(t *testing.T) {
	w, tools := gloutonWorld(t)
	if _, err := runTool(t, tools, "glouton_save_draft", `{"personId":"p1","subject":"Hello","text":"Hi"}`); err != nil {
		t.Fatal(err)
	}
	if _, err := runTool(t, tools, "glouton_create_person", `{"firstName":"Ann","companyId":"c1"}`); err != nil {
		t.Fatal(err)
	}
	if len(w.writes) != 2 || !strings.Contains(w.writes[0].body, `"personId":"p1"`) ||
		!strings.Contains(w.writes[1].body, `"firstName":"Ann"`) || !strings.Contains(w.writes[1].body, `"companyId":"c1"`) {
		t.Fatalf("old arguments must decode and Glouton must still get camelCase: %+v", w.writes)
	}
}
