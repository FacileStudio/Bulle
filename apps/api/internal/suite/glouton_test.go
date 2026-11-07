package suite

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/FacileStudio/nacelle"
)

func serveGlouton(rw http.ResponseWriter, r *http.Request, user string) {
	switch {
	case r.Method != http.MethodGet:
		body, _ := io.ReadAll(r.Body)
		_, _ = fmt.Fprintf(rw, `{"data":{"id":"new","echo":%s}}`, body)
	case r.URL.Path == gloutonContactsPath:
		_, _ = fmt.Fprintf(rw, `{"data":[{"type":"company","id":"c1","name":"prospect of %s","query":%q}]}`, user, r.URL.RawQuery)
	default:
		_, _ = fmt.Fprintf(rw, `{"data":{"path":%q,"query":%q}}`, r.URL.Path, r.URL.RawQuery)
	}
}

func gloutonWorld(t *testing.T) (*world, []nacelle.Tool) {
	t.Helper()
	w := newWorld(t)
	w.addUser(1, future)
	tools, _, err := Tools(NewGateway(w.tokens(), w.cfg, nil), 1, Defaults{}, Runtime{})
	if err != nil {
		t.Fatal(err)
	}
	return w, tools
}

func TestGloutonReadsUnwrapAndFilter(t *testing.T) {
	w, tools := gloutonWorld(t)
	out, err := runTool(t, tools, "glouton_search_contacts", `{"query":"bakery","temperature":"HOT","limit":500}`)
	if err != nil || !strings.HasPrefix(out, "[") || !strings.Contains(out, "prospect of 1") {
		t.Fatalf("contacts must come back unwrapped for this user: %q, %v", out, err)
	}
	for _, want := range []string{"q=bakery", "temperature=HOT", "limit=100"} {
		if !strings.Contains(out, want) {
			t.Errorf("contacts query lost %s: %s", want, out)
		}
	}
	company, _ := runTool(t, tools, "glouton_get_company", `{"company_id":"c1"}`)
	suggestions, _ := runTool(t, tools, "glouton_list_suggestions", `{}`)
	if !strings.Contains(company, `"path":"/api/agent/companies/c1"`) {
		t.Errorf("company path: %s", company)
	}
	if !strings.Contains(suggestions, "status=OPEN") {
		t.Errorf("suggestions must default to open ones: %s", suggestions)
	}
	if _, err := runTool(t, tools, "glouton_list_outreach", `{}`); err == nil {
		t.Error("outreach without a company or person must be refused")
	}
	if len(w.writes) != 0 {
		t.Fatalf("reads must not write: %+v", w.writes)
	}
}

func TestGloutonWritesSendOnlyWhatChanges(t *testing.T) {
	w, tools := gloutonWorld(t)
	calls := []struct{ tool, input, method, path string }{
		{"glouton_save_draft", `{"person_id":"p1","subject":"Hello","text":"Hi","rationale":"new HOT lead"}`, http.MethodPost, gloutonDraftsPath},
		{"glouton_update_draft", `{"draft_id":"d1","subject":"Hello again"}`, http.MethodPatch, gloutonDraftsPath + "/d1"},
		{"glouton_update_company", `{"company_id":"c1","score":80,"temperature":"HOT"}`, http.MethodPatch, gloutonCompaniesPath + "/c1"},
		{"glouton_suggest", `{"kind":"RESCORE","title":"Warm up","company_id":"c1","temperature":"WARM","score":55}`, http.MethodPost, gloutonSuggestionsPath},
	}
	for _, call := range calls {
		if _, err := runTool(t, tools, call.tool, call.input); err != nil {
			t.Fatalf("%s: %v", call.tool, err)
		}
	}
	if len(w.writes) != len(calls) {
		t.Fatalf("want %d writes, got %+v", len(calls), w.writes)
	}
	bodies := make([]map[string]any, len(calls))
	for i, call := range calls {
		got := w.writes[i]
		if got.method != call.method || got.path != call.path {
			t.Errorf("%s: %s %s, want %s %s", call.tool, got.method, got.path, call.method, call.path)
		}
		_ = json.Unmarshal([]byte(got.body), &bodies[i])
	}
	if bodies[0]["personId"] != "p1" || bodies[0]["text"] != "Hi" || bodies[0]["html"] != nil {
		t.Errorf("draft body: %v", bodies[0])
	}
	if len(bodies[1]) != 1 || len(bodies[2]) != 2 || bodies[2]["score"] != float64(80) {
		t.Errorf("updates must carry only the passed fields: %v | %v", bodies[1], bodies[2])
	}
	payload, _ := bodies[3]["payload"].(map[string]any)
	if payload["temperature"] != "WARM" || payload["score"] != float64(55) || bodies[3]["temperature"] != nil {
		t.Errorf("rescore must travel in the payload: %v", bodies[3])
	}
}

func TestGloutonHasNoSendConvertOrDelete(t *testing.T) {
	w, tools := gloutonWorld(t)
	forbidden := regexp.MustCompile(`^glouton_.*(send|convert|delete|merge)`)
	inbox := w.cfg.Apps[AppGlouton].URL + "/app/inbox"
	for _, tool := range tools {
		if forbidden.MatchString(tool.Spec().Name) {
			t.Errorf("%s must not exist: only a human sends from Glouton", tool.Spec().Name)
		}
		if strings.HasPrefix(tool.Spec().Name, "glouton_") && strings.Contains(tool.Spec().Name, "draft") && !strings.Contains(tool.Spec().Description, inbox) {
			t.Errorf("%s must point the model at %s: %q", tool.Spec().Name, inbox, tool.Spec().Description)
		}
	}
	if _, err := runTool(t, tools, "glouton_save_draft", `{"person_id":"p1","text":"no subject"}`); err == nil || len(w.writes) != 0 {
		t.Fatalf("an incomplete draft must be refused before any request: %v", err)
	}
}

func TestGloutonStatusProbe(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	for _, name := range appNames {
		if name != AppGlouton {
			w.cfg.Apps[name] = AppConfig{}
		}
	}
	got := NewGateway(w.tokens(), w.cfg, nil).Statuses(t.Context(), 1)
	if len(got) != 1 || got[0].App != AppGlouton || got[0].Name != "Glouton" || got[0].State != StateConnected {
		t.Fatalf("want Glouton connected, got %+v", got)
	}
}
