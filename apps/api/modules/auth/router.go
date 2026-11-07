// Package auth serves GET /auth/me and provisions a new account, with its
// starter jobs, on first sign-in.
package auth

import (
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service, st store.Store, plan PlanFunc, requireAuth func(http.Handler) http.Handler) {
	handler := &Handler{service: service, store: st, plan: plan}
	router.With(requireAuth).Get("/auth/me", handler.me)
}
