// Package assistant serves the routes that read and save the user's assistant
// profile.
package assistant

import (
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, st store.Store, requireAuth func(http.Handler) http.Handler) {
	handler := NewHandler(st)
	router.With(requireAuth).Get("/assistant", handler.getProfile)
	router.With(requireAuth).Put("/assistant", handler.saveProfile)
}
