package preferences

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/catalog"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/go-chi/chi/v5"
)

func testAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(authcontext.WithIdentity(r.Context(), authcontext.Identity{UserID: 7})))
	})
}

func testCatalog(_ context.Context, _ int64) []catalog.Model {
	return []catalog.Model{{ID: "model-a", Name: "A"}, {ID: "vendor/b", Name: "B"}}
}

func setup(t *testing.T) (*httptest.Server, store.Store) {
	t.Helper()
	st := store.NewMemory()
	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		RegisterRoutes(r, st, testCatalog, "model-a", testAuth)
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
	if prefs.Model != "" || prefs.DefaultModel != "model-a" || len(prefs.Catalog) != 2 {
		t.Fatalf("fresh preferences = %+v, want no model over the instance default", prefs)
	}

	status, raw := call(t, srv, http.MethodPut, "/api/preferences", `{"model":"vendor/b"}`)
	if status != http.StatusOK {
		t.Fatalf("PUT model = %d: %s", status, raw)
	}
	if got := readPrefs(t, srv); got.Model != "vendor/b" {
		t.Fatalf("model = %q, want vendor/b", got.Model)
	}
}

func TestPreferencesClearRestoresTheInstanceDefault(t *testing.T) {
	srv, _ := setup(t)
	if status, raw := call(t, srv, http.MethodPut, "/api/preferences", `{"model":"vendor/b"}`); status != http.StatusOK {
		t.Fatalf("PUT = %d: %s", status, raw)
	}
	if status, raw := call(t, srv, http.MethodPut, "/api/preferences", `{"model":""}`); status != http.StatusOK {
		t.Fatalf("clear = %d: %s", status, raw)
	}
	if got := readPrefs(t, srv); got.Model != "" {
		t.Fatalf("model = %q after clearing, want the instance default", got.Model)
	}
}

func TestPreferencesRefuseAModelTheBackendDoesNotOffer(t *testing.T) {
	srv, _ := setup(t)
	if status, raw := call(t, srv, http.MethodPut, "/api/preferences", `{"model":"model-z"}`); status != http.StatusBadRequest {
		t.Fatalf("status = %d (%s), want 400", status, raw)
	}
}
