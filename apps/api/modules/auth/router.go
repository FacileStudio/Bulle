package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service, requireAuth func(http.Handler) http.Handler) {
	handler := &Handler{service: service}
	router.With(requireAuth).Get("/auth/me", handler.me)
}
