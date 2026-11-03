package spaces

import (
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, st store.Store, requireAuth func(http.Handler) http.Handler) {
	handler := NewHandler(st)
	router.With(requireAuth).Get("/spaces", handler.listSpaces)
	router.With(requireAuth).Post("/spaces", handler.createSpace)
	router.With(requireAuth).Put("/spaces/{id}", handler.updateSpace)
	router.With(requireAuth).Delete("/spaces/{id}", handler.deleteSpace)
}
