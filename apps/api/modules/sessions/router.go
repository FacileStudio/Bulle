// Package sessions serves the chat routes: sessions, their messages, runs, live
// events and approvals.
package sessions

import (
	"log/slog"
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, st store.Store, rn *runner.Runner, logger *slog.Logger, requireAuth func(http.Handler) http.Handler) {
	handler := NewHandler(st, rn, logger)
	router.With(requireAuth).Get("/sessions", handler.listSessions)
	router.With(requireAuth).Post("/sessions", handler.createSession)
	router.With(requireAuth).Get("/sessions/search", handler.searchSessions)
	router.With(requireAuth).Get("/sessions/{id}", handler.getSession)
	router.With(requireAuth).Delete("/sessions/{id}", handler.deleteSession)
	router.With(requireAuth).Post("/sessions/{id}/messages", handler.sendMessage)
	router.With(requireAuth).Get("/runs/{id}/events", handler.streamEvents)
	router.With(requireAuth).Post("/runs/{id}/stop", handler.stopRun)
	router.With(requireAuth).Post("/runs/{id}/approvals/{call}", handler.decideApproval)
}
