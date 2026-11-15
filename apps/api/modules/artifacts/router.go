// Package artifacts serves the routes that list, search, read and delete a space's
// Mycelium artifacts.
package artifacts

import (
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	troncmiddleware "github.com/FacileStudio/tronc/middleware"
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts the artifact routes. A nil gateway is allowed: the
// list then answers available false. Public links are served only when
// sharing has a store. Their public read is rate limited per client address.
func RegisterRoutes(router chi.Router, st store.Store, gateway *suite.Gateway, sharing Sharing, requireAuth func(http.Handler) http.Handler) {
	handler := &Handler{store: st, gateway: gateway, sharing: sharing}
	router.With(requireAuth).Get("/artifacts", handler.listArtifacts)
	router.With(requireAuth).Get("/artifacts/search", handler.searchArtifacts)
	router.With(requireAuth).Get("/artifacts/{id}", handler.getArtifact)
	router.With(requireAuth).Put("/artifacts/{id}", handler.updateArtifact)
	router.With(requireAuth).Delete("/artifacts/{id}", handler.deleteArtifact)
	if sharing.Store == nil {
		return
	}
	publicLimit := troncmiddleware.Ratelimit(troncmiddleware.RatelimitConfig{Limit: 2, Burst: 30})
	router.With(requireAuth).Post("/artifacts/{id}/share", handler.createShare)
	router.With(requireAuth).Get("/artifacts/{id}/share", handler.getShare)
	router.With(requireAuth).Delete("/artifacts/{id}/share", handler.revokeShare)
	router.With(publicLimit).Get("/public/shares/{token}", handler.publicShare)
}
