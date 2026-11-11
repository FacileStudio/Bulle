// Package ai serves POST /ai/decide, POST /ai/summarize and POST /ai/complete,
// the AI other suite services call so none of them holds an AI key.
package ai

import (
	"log/slog"
	"time"

	troncmiddleware "github.com/FacileStudio/tronc/middleware"
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts POST /ai/decide and POST /ai/summarize, which only a
// service named in Callers may call, never a browser session or a user token,
// each with its own rate limit. With RequireAuth set it also mounts
// POST /ai/complete, which takes a token an app exchanged for its user.
func RegisterRoutes(router chi.Router, cfg Config) {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.Decider == nil && len(cfg.Callers) > 0 {
		cfg.Logger.Warn("BULLE_AI_CALLERS is set but no jev key is: /api/ai/decide answers 503")
	}
	handler := &handler{cfg: cfg}
	limit := troncmiddleware.Ratelimit(troncmiddleware.RatelimitConfig{Limit: 2, Burst: 30, KeyFunc: callerOf})
	router.With(handler.authenticate, limit).Post("/ai/decide", handler.decide)
	perUser := troncmiddleware.Ratelimit(troncmiddleware.RatelimitConfig{Limit: 0.2, Burst: 10, KeyFunc: callerAndUser})
	router.With(handler.authenticate, perUser).Post("/ai/summarize", handler.summarize)
	if cfg.RequireAuth != nil {
		completions := &completion{cfg: cfg, cap: &dailyCap{limit: CompletionsPerDay, now: time.Now}}
		perActor := troncmiddleware.Ratelimit(troncmiddleware.RatelimitConfig{Limit: 0.2, Burst: 10, KeyFunc: actorAndUser})
		router.With(cfg.RequireAuth, perActor).Post("/ai/complete", completions.complete)
	}
}
