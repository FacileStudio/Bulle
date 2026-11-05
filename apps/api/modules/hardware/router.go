package hardware

import (
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	troncmiddleware "github.com/FacileStudio/tronc/middleware"
	"github.com/go-chi/chi/v5"
)

const defaultKeepalive = 30 * time.Second

// RegisterRoutes mounts device pairing, listing, revocation and the bohs.v1
// gateway socket. The socket authenticates on its own, since a device token
// is not a session. Claims are rate limited per user and socket upgrades per
// client address.
func RegisterRoutes(router chi.Router, cfg Config, requireAuth func(http.Handler) http.Handler) {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.Keepalive <= 0 {
		cfg.Keepalive = defaultKeepalive
	}
	cfg.Origins = slices.DeleteFunc(slices.Clone(cfg.Origins), func(origin string) bool { return origin == "*" })
	handler := &Handler{cfg: cfg, requireAuth: requireAuth}
	claimLimit := troncmiddleware.Ratelimit(troncmiddleware.RatelimitConfig{Limit: 1.0 / 12, Burst: 5, KeyFunc: userKey})
	socketLimit := troncmiddleware.Ratelimit(troncmiddleware.RatelimitConfig{Limit: 1, Burst: 20})
	router.With(requireAuth).Get("/hardware/devices", handler.list)
	router.With(requireAuth, claimLimit).Post("/hardware/devices/claim", handler.claim)
	router.With(requireAuth).Delete("/hardware/devices/{id}", handler.revoke)
	router.With(socketLimit).Get("/hardware/ws", handler.socket)
}

func userKey(r *http.Request) string {
	identity, _ := authcontext.IdentityFromContext(r.Context())
	return strconv.FormatInt(identity.UserID, 10)
}
