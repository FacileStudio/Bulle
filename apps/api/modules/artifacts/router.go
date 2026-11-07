// Package artifacts serves the routes that list, read and delete a space's
// Mycelium artifacts.
package artifacts

import (
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts the artifact routes. A nil gateway is allowed: the
// list then answers available false.
func RegisterRoutes(router chi.Router, st store.Store, gateway *suite.Gateway, requireAuth func(http.Handler) http.Handler) {
	handler := &Handler{store: st, gateway: gateway}
	router.With(requireAuth).Get("/artifacts", handler.listArtifacts)
	router.With(requireAuth).Get("/artifacts/{id}", handler.getArtifact)
	router.With(requireAuth).Delete("/artifacts/{id}", handler.deleteArtifact)
}
