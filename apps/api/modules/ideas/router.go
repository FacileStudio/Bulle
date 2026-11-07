// Package ideas serves the Ideas page: the idea cards, goals and daily batch
// settings of a space.
package ideas

import (
	"log/slog"
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, st store.Store, rn *runner.Runner, logger *slog.Logger, requireAuth func(http.Handler) http.Handler) {
	handler := NewHandler(st, rn, logger)
	router.With(requireAuth).Get("/ideas", handler.list)
	router.With(requireAuth).Post("/ideas/generate", handler.generate)
	router.With(requireAuth).Put("/ideas/settings", handler.saveSettings)
	router.With(requireAuth).Post("/ideas/goals", handler.createGoal)
	router.With(requireAuth).Put("/ideas/goals/{id}", handler.updateGoal)
	router.With(requireAuth).Delete("/ideas/goals/{id}", handler.deleteGoal)
	router.With(requireAuth).Patch("/ideas/{id}", handler.update)
	router.With(requireAuth).Post("/ideas/{id}/start", handler.start)
}
