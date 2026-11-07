// Package preferences serves the routes that read and save the user's default
// model.
package preferences

import (
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, st store.Store, catalog Catalog, defaultModel string, requireAuth func(http.Handler) http.Handler) {
	handler := NewHandler(st, catalog, defaultModel)
	router.With(requireAuth).Get("/preferences", handler.Get)
	router.With(requireAuth).Put("/preferences", handler.Update)
}
