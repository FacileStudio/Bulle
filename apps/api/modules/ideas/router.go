// Package ideas serves the Ideas page: the idea cards, goals and daily batch
// settings of a space, and the per-app API other suite apps call to show and
// run ideas in place.
package ideas

import (
	"log/slog"
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/porte"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, st store.Store, rn *runner.Runner, logger *slog.Logger, requireAuth func(http.Handler) http.Handler) {
	handler := NewHandler(st, rn, logger)
	auth := router.With(requireAuth, actingForUser)
	auth.Get("/ideas", handler.list)
	auth.Post("/ideas/generate", handler.generate)
	auth.Put("/ideas/settings", handler.saveSettings)
	auth.Post("/ideas/goals", handler.createGoal)
	auth.Put("/ideas/goals/{id}", handler.updateGoal)
	auth.Delete("/ideas/goals/{id}", handler.deleteGoal)
	auth.Patch("/ideas/{id}", handler.update)
	auth.Post("/ideas/{id}/start", handler.start)
	auth.Post("/ideas/{id}/run", handler.run)
	auth.Get("/runs/{id}", handler.getRun)
}

// actingForUser refuses a bearer JWT that does not act for its subject. Apps
// reach these routes with an RFC 8693 token exchanged for audience bulle,
// whose act claim names the app and whose sub is the user; porte has already
// checked the audience and the bulle.read or bulle.write scope. A JWT with no
// act is a service account speaking for nobody.
func actingForUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if identity, ok := porte.From(r.Context()); ok && identity.Subject != "" && identity.Actor == "" {
			httpjson.WriteError(w, troncerrors.Forbidden("a bearer token must act for a user"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
