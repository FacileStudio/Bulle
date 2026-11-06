package artifacts_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/Bulle/apps/api/modules/artifacts"
	"github.com/go-chi/chi/v5"
)

func router(gateway *suite.Gateway, userID int64) http.Handler {
	r := chi.NewRouter()
	requireAuth := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if userID > 0 {
				req = req.WithContext(authcontext.WithIdentity(req.Context(), authcontext.Identity{UserID: userID}))
			}
			next.ServeHTTP(w, req)
		})
	}
	artifacts.RegisterRoutes(r, store.NewMemory(), gateway, requireAuth)
	return r
}

func call(h http.Handler, method, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

func TestArtifactsAreUnavailableWithoutMycelium(t *testing.T) {
	noMycelium := suite.NewGateway(nil, suite.Config{Apps: map[string]suite.AppConfig{
		suite.AppCourrier: {URL: "http://courrier", Audience: "courrier"},
	}}, nil)
	for name, gateway := range map[string]*suite.Gateway{"no suite": nil, "no mycelium": noMycelium} {
		h := router(gateway, 1)
		if rec := call(h, http.MethodGet, "/artifacts?space_id=s"); rec.Code != http.StatusOK || rec.Body.String() != "{\"artifacts\":[],\"available\":false}\n" {
			t.Errorf("%s: list = %d %s", name, rec.Code, rec.Body)
		}
		if rec := call(h, http.MethodGet, "/artifacts/a"); rec.Code != http.StatusPreconditionFailed {
			t.Errorf("%s: get = %d", name, rec.Code)
		}
		if rec := call(h, http.MethodDelete, "/artifacts/a"); rec.Code != http.StatusPreconditionFailed {
			t.Errorf("%s: delete = %d", name, rec.Code)
		}
	}
	if rec := call(router(nil, 0), http.MethodGet, "/artifacts"); rec.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous list = %d", rec.Code)
	}
}
