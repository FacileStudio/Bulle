package memory_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/Bulle/apps/api/modules/memory"
	"github.com/go-chi/chi/v5"
)

func router(st store.Store, gateway *suite.Gateway, userID int64) http.Handler {
	r := chi.NewRouter()
	requireAuth := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if userID > 0 {
				req = req.WithContext(authcontext.WithIdentity(req.Context(), authcontext.Identity{UserID: userID}))
			}
			next.ServeHTTP(w, req)
		})
	}
	memory.RegisterRoutes(r, st, gateway, requireAuth)
	return r
}

func call(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
	return rec
}

func myceliumGateway(t *testing.T) (*suite.Gateway, *atomic.Int32) {
	t.Helper()
	hits := &atomic.Int32{}
	app := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusTeapot)
	}))
	t.Cleanup(app.Close)
	cfg := suite.Config{MCP: map[string]suite.MCPConfig{suite.AppMycelium: {URL: app.URL + "/mcp", Resource: app.URL + "/mcp"}}}
	return suite.NewGateway(nil, cfg, nil), hits
}

func TestMemoryIsUnavailableWithoutMycelium(t *testing.T) {
	h := router(store.NewMemory(), nil, 1)
	if rec := call(h, http.MethodGet, "/memory/pages", ""); rec.Code != http.StatusOK || rec.Body.String() != "{\"pages\":[],\"available\":false}\n" {
		t.Fatalf("list = %d %s", rec.Code, rec.Body)
	}
	for _, method := range []string{http.MethodGet, http.MethodDelete} {
		if rec := call(h, method, "/memory/pages/profile.md", ""); rec.Code != http.StatusPreconditionFailed {
			t.Errorf("%s = %d", method, rec.Code)
		}
	}
	if rec := call(h, http.MethodPut, "/memory/pages/profile.md", `{"content":"hi"}`); rec.Code != http.StatusPreconditionFailed {
		t.Errorf("put = %d", rec.Code)
	}
	if rec := call(router(store.NewMemory(), nil, 0), http.MethodGet, "/memory/pages", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous list = %d", rec.Code)
	}
}

func TestAnotherUsersSpaceIsRefused(t *testing.T) {
	st := store.NewMemory()
	theirs, err := st.CreateSpace(context.Background(), store.Space{UserID: 1, Name: "Theirs", MyceliumSpaceID: "sp-1"})
	if err != nil {
		t.Fatal(err)
	}
	gateway, hits := myceliumGateway(t)
	h := router(st, gateway, 2)
	query := "?space_id=" + theirs.ID
	for _, route := range [][3]string{
		{http.MethodGet, "/memory/pages" + query, ""},
		{http.MethodGet, "/memory/pages/profile.md" + query, ""},
		{http.MethodPut, "/memory/pages/profile.md" + query, `{"content":"x"}`},
		{http.MethodDelete, "/memory/pages/profile.md" + query, ""},
	} {
		if rec := call(h, route[0], route[1], route[2]); rec.Code != http.StatusNotFound {
			t.Errorf("%s %s = %d", route[0], route[1], rec.Code)
		}
	}
	if hits.Load() != 0 {
		t.Fatalf("Mycelium was reached %d times", hits.Load())
	}
}

func TestPagePathsAreChecked(t *testing.T) {
	gateway, hits := myceliumGateway(t)
	h := router(store.NewMemory(), gateway, 1)
	for _, path := range []string{"..%2Fsecrets.md", "a/../../b.md", ".hidden.md", "people/.x.md", "notes.txt", "a%5Cb.md", "%2Fetc%2Fx.md"} {
		for _, method := range []string{http.MethodGet, http.MethodDelete, http.MethodPut} {
			if rec := call(h, method, "/memory/pages/"+path, `{"content":"x"}`); rec.Code != http.StatusBadRequest {
				t.Errorf("%s %s = %d", method, path, rec.Code)
			}
		}
	}
	for _, body := range []string{`{"content":"  "}`, `{"content":"` + strings.Repeat("a", 65<<10) + `"}`, `{"nope":1}`} {
		if rec := call(h, http.MethodPut, "/memory/pages/a.md", body); rec.Code != http.StatusBadRequest {
			t.Errorf("put body of %d bytes = %d", len(body), rec.Code)
		}
	}
	if hits.Load() != 0 {
		t.Fatalf("a refused request reached Mycelium %d times", hits.Load())
	}
}
