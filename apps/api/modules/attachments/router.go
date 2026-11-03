package attachments

import (
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, st store.Attachments, requireAuth func(http.Handler) http.Handler) {
	handler := NewHandler(st)
	router.With(requireAuth).Post("/attachments", handler.Upload)
	router.With(requireAuth).Get("/attachments/{id}", handler.Download)
}
