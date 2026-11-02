package auth

import (
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service, st store.Store, requireAuth func(http.Handler) http.Handler) {
	handler := &Handler{service: service, store: st}
	router.With(requireAuth).Get("/auth/me", handler.me)
}
