// Package jobs serves the routes that create, edit, run and list scheduled jobs.
package jobs

import (
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, st store.Store, rn *runner.Runner, requireAuth func(http.Handler) http.Handler) {
	handler := NewHandler(st, rn)
	router.With(requireAuth).Get("/jobs", handler.listJobs)
	router.With(requireAuth).Post("/jobs", handler.createJob)
	router.With(requireAuth).Get("/jobs/{id}", handler.getJob)
	router.With(requireAuth).Put("/jobs/{id}", handler.updateJob)
	router.With(requireAuth).Delete("/jobs/{id}", handler.deleteJob)
	router.With(requireAuth).Post("/jobs/{id}/run", handler.runJob)
	router.With(requireAuth).Get("/jobs/{id}/runs", handler.listJobRuns)
}
