package push

import (
	"net/http"

	pushsender "github.com/FacileStudio/Bulle/apps/api/internal/push"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, hub *pushsender.Hub, requireAuth func(http.Handler) http.Handler) {
	handler := NewHandler(hub)
	router.With(requireAuth).Get("/push", handler.status)
	router.With(requireAuth).Post("/push/test", handler.test)
	router.With(requireAuth).Get("/push/stream", handler.stream)
	router.With(requireAuth).Post("/push/devices", handler.registerDevice)
	router.With(requireAuth).Delete("/push/devices", handler.forgetDevice)
}
