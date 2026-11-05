package preferences

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/go-chi/chi/v5"
)

func testAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(authcontext.WithIdentity(r.Context(), authcontext.Identity{UserID: 7})))
	})
}

func setup(t *testing.T) (*httptest.Server, store.Store) {
	t.Helper()
	st := store.NewMemory()
	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		RegisterRoutes(r, st, []string{"model-a", "model-b"}, "model-a", testAuth)
	})
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	return srv, st
}

func call(t *testing.T, srv *httptest.Server, method, path, body string) (int, []byte) {
	t.Helper()
	req, err := http.NewRequest(method, srv.URL+path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, raw
}

func readPrefs(t *testing.T, srv *httptest.Server) PreferencesResponse {
	t.Helper()
	status, raw := call(t, srv, http.MethodGet, "/api/preferences", "")
	if status != http.StatusOK {
		t.Fatalf("GET /preferences = %d: %s", status, raw)
	}
	var out PreferencesResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestPreferencesDefaultsAndRoundTrip(t *testing.T) {
	srv, _ := setup(t)
	prefs := readPrefs(t, srv)
	if prefs.Model != "" || prefs.DefaultModel != "model-a" || len(prefs.Models) != 2 {
		t.Fatalf("fresh preferences = %+v, want no model over the instance default", prefs)
	}

	status, raw := call(t, srv, http.MethodPut, "/api/preferences", `{"model":"model-b"}`)
	if status != http.StatusOK {
		t.Fatalf("PUT model = %d: %s", status, raw)
	}
	if got := readPrefs(t, srv); got.Model != "model-b" {
		t.Fatalf("model = %q, want model-b", got.Model)
	}
}

func TestPreferencesKeepTheFieldTheRequestLeavesOut(t *testing.T) {
	srv, _ := setup(t)
	if status, raw := call(t, srv, http.MethodPut, "/api/preferences", `{"model":"model-b"}`); status != http.StatusOK {
		t.Fatalf("PUT model = %d: %s", status, raw)
	}
	if status, raw := call(t, srv, http.MethodPut, "/api/preferences", `{"cli_config":"limits:\n  max_concurrency: 4\n"}`); status != http.StatusOK {
		t.Fatalf("PUT cli_config = %d: %s", status, raw)
	}
	prefs := readPrefs(t, srv)
	if prefs.Model != "model-b" {
		t.Errorf("model = %q after a CLI-only save, want model-b kept", prefs.Model)
	}
	if !strings.Contains(prefs.CLIConfig, "max_concurrency: 4") {
		t.Errorf("cli_config = %q, want the saved document", prefs.CLIConfig)
	}

	status, raw := call(t, srv, http.MethodGet, "/api/cli/config", "")
	if status != http.StatusOK || !strings.Contains(string(raw), "max_concurrency: 4") {
		t.Fatalf("GET /cli/config = %d: %s", status, raw)
	}
}

func TestPreferencesRefuseWhatTheyCannotHonour(t *testing.T) {
	srv, _ := setup(t)
	cases := map[string]string{
		"a model the instance does not offer": `{"model":"model-z"}`,
		"an unknown config key":               `{"cli_config":"limits:\n  max_iteration: 5\n"}`,
		"a config that moves the server":      `{"cli_config":"server:\n  url: https://elsewhere.example\n"}`,
	}
	for name, body := range cases {
		if status, raw := call(t, srv, http.MethodPut, "/api/preferences", body); status != http.StatusBadRequest {
			t.Errorf("%s: status = %d (%s), want 400", name, status, raw)
		}
	}
}
