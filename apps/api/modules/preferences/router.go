package preferences

import (
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, st store.Store, models []string, defaultModel string, requireAuth func(http.Handler) http.Handler) {
	handler := NewHandler(st, models, defaultModel)
	router.With(requireAuth).Get("/preferences", handler.Get)
	router.With(requireAuth).Put("/preferences", handler.Update)
	router.With(requireAuth).Get("/cli/config", handler.CLIConfig)
}
