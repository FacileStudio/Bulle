// Package suite serves GET /suite/status, which reports the suite apps the user
// is connected to.
package suite

import (
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, gateway *suite.Gateway, requireAuth func(http.Handler) http.Handler) {
	handler := &Handler{gateway: gateway}
	router.With(requireAuth).Get("/suite/status", handler.status)
}
