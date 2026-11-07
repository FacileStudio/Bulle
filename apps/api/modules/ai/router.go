package ai

import (
	"log/slog"

	troncmiddleware "github.com/FacileStudio/tronc/middleware"
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts POST /ai/decide. Only a service named in Callers may
// call it, never a browser session or a user token, and each service has its
// own rate limit.
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
}
