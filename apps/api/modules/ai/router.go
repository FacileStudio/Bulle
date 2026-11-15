// Package ai serves POST /ai/decide, POST /ai/summarize and POST /ai/complete,
// the AI other suite services call so none of them holds an AI key.
package ai

import (
	"log/slog"
	"time"

	troncmiddleware "github.com/FacileStudio/tronc/middleware"
	"github.com/go-chi/chi/v5"
)

// limiterKeys bounds how many caller and user pairs each limiter remembers.
const limiterKeys = 10_000

// RegisterRoutes mounts POST /ai/decide and POST /ai/summarize, which take a
// token an app exchanged for its user (never a browser session or a plain user
// token). Each checks the plan
// of the user the call acts for first, then spends a rate limit kept per app
// and user. With RequireAuth set it also mounts POST /ai/complete.
func RegisterRoutes(router chi.Router, cfg Config) {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.Decider == nil && cfg.RequireAuth != nil {
		cfg.Logger.Warn("no jev key is set: /api/ai/decide answers 503")
	}
	handler := &handler{cfg: cfg}
	decisions := troncmiddleware.Ratelimit(troncmiddleware.RatelimitConfig{Limit: 2, Burst: 30, KeyFunc: callerAndUser, MaxKeys: limiterKeys})
	router.With(handler.authenticate, handler.admit, decisions).Post("/ai/decide", handler.decide)
	summaries := troncmiddleware.Ratelimit(troncmiddleware.RatelimitConfig{Limit: 0.2, Burst: 10, KeyFunc: callerAndUser, MaxKeys: limiterKeys})
	router.With(handler.authenticate, handler.admit, summaries).Post("/ai/summarize", handler.summarize)
	if cfg.RequireAuth != nil {
		completionCap := cfg.Cap
		if completionCap == nil {
			completionCap = &dailyCap{limit: CompletionsPerDay, now: time.Now}
		}
		completions := &completion{cfg: cfg, cap: completionCap}
		perActor := troncmiddleware.Ratelimit(troncmiddleware.RatelimitConfig{Limit: 0.2, Burst: 10, KeyFunc: actorAndUser, MaxKeys: limiterKeys})
		router.With(cfg.RequireAuth, perActor).Post("/ai/complete", completions.complete)
	}
}
