// Package credentials serves the routes that list, save, test and remove the
// user's own API keys.
package credentials

import (
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/credentials"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, store credentials.Store, requireAuth func(http.Handler) http.Handler) {
	handler := NewHandler(store, nil)
	router.With(requireAuth).Get("/credentials", handler.List)
	router.With(requireAuth).Put("/credentials/{backend}", handler.Put)
	router.With(requireAuth).Delete("/credentials/{backend}", handler.Remove)
	router.With(requireAuth).Post("/credentials/{backend}/test", handler.Test)
}
