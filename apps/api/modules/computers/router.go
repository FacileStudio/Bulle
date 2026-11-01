package computers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, manager Manager, requireAuth func(http.Handler) http.Handler) {
	handler := &Handler{manager: manager}
	router.With(requireAuth).Get("/computer", handler.get)
	router.With(requireAuth).Post("/computer/start", handler.start)
	router.With(requireAuth).Post("/computer/stop", handler.stop)
	router.With(requireAuth).Post("/computer/snapshots", handler.snapshot)
	router.With(requireAuth).Post("/computer/rollback", handler.rollback)
}
