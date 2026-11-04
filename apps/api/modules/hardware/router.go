package hardware

import (
	"log/slog"
	"net/http"
	"slices"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts device pairing, listing, revocation and the bohs.v1
// gateway socket. The socket authenticates on its own, since a device token
// is not a session.
func RegisterRoutes(router chi.Router, cfg Config, requireAuth func(http.Handler) http.Handler) {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	cfg.Origins = slices.DeleteFunc(slices.Clone(cfg.Origins), func(origin string) bool { return origin == "*" })
	handler := &Handler{cfg: cfg, requireAuth: requireAuth}
	router.With(requireAuth).Get("/hardware/devices", handler.list)
	router.With(requireAuth).Post("/hardware/devices/claim", handler.claim)
	router.With(requireAuth).Delete("/hardware/devices/{id}", handler.revoke)
	router.Get("/hardware/ws", handler.socket)
}
