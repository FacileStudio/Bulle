// Package memory serves the routes that list, read, edit and delete a space's
// Mycelium memory pages.
package memory

import (
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts the memory page routes. A nil gateway is allowed: the
// list then answers available false.
func RegisterRoutes(router chi.Router, st store.Store, gateway *suite.Gateway, requireAuth func(http.Handler) http.Handler) {
	handler := &Handler{store: st, gateway: gateway}
	router.With(requireAuth).Get("/memory/pages", handler.list)
	router.With(requireAuth).Get("/memory/pages/*", handler.read)
	router.With(requireAuth).Put("/memory/pages/*", handler.write)
	router.With(requireAuth).Delete("/memory/pages/*", handler.remove)
}
