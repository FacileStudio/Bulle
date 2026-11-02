package push

import (
	"net/http"

	pushsender "github.com/FacileStudio/Bulle/apps/api/internal/push"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, sender *pushsender.Sender, requireAuth func(http.Handler) http.Handler) {
	handler := NewHandler(sender)
	router.With(requireAuth).Get("/push", handler.status)
	router.With(requireAuth).Post("/push/test", handler.test)
}
