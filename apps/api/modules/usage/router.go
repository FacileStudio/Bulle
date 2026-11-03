package usage

import (
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts the usage summary.
func RegisterRoutes(router chi.Router, st store.Store, defaultModel string, requireAuth func(http.Handler) http.Handler) {
	router.With(requireAuth).Get("/usage", NewHandler(st, defaultModel).summary)
}
