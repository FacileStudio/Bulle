package main

import (
	"context"
	"log/slog"
	"net/http"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/middleware"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	memorymodule "github.com/FacileStudio/Bulle/apps/api/modules/memory"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/session"
	"github.com/go-chi/chi/v5"
)

func TestMemoryCookieMutationsNeedCSRF(t *testing.T) {
	manager, err := session.New(porte.Config{SuccessURL: "http://bulle.test/"}, session.Deps{
		Sessions: newMemoryStore(),
		Logger:   slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatal(err)
	}
	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		memorymodule.RegisterRoutes(r, store.NewMemory(), nil, middleware.RequireAuth(manager, fixedResolver{}))
	})
	token, _, err := manager.Issue(context.Background(), 7, "")
	if err != nil {
		t.Fatal(err)
	}
	cookie := map[string]string{"Cookie": porte.SessionCookieName + "=" + token}
	if code := doJSON(router, http.MethodGet, "/api/memory/pages", "", cookie).Code; code != http.StatusOK {
		t.Fatalf("cookie GET = %d", code)
	}
	for _, method := range []string{http.MethodPut, http.MethodDelete} {
		if code := doJSON(router, method, "/api/memory/pages/profile.md", `{"content":"x"}`, cookie).Code; code != http.StatusForbidden {
			t.Errorf("cookie %s without CSRF = %d", method, code)
		}
	}
	cookie[porte.CSRFHeaderName] = "1"
	for _, method := range []string{http.MethodPut, http.MethodDelete} {
		if code := doJSON(router, method, "/api/memory/pages/profile.md", `{"content":"x"}`, cookie).Code; code != http.StatusPreconditionFailed {
			t.Errorf("cookie %s with CSRF = %d, want the no-Mycelium 412", method, code)
		}
	}
}
