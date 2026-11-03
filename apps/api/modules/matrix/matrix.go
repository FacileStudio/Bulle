// Package matrix serves the status of the user's chat agent: its Matrix
// account in Prisme, the rooms it sits in and whether it is running.
package matrix

import (
	"context"
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/matrixbot"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

// Bots is the part of the bot manager these routes use.
type Bots interface {
	Get(ctx context.Context, userID int64) (matrixbot.Status, error)
}

// RegisterRoutes mounts GET /matrix.
func RegisterRoutes(router chi.Router, bots Bots, requireAuth func(http.Handler) http.Handler) {
	router.With(requireAuth).Get("/matrix", func(w http.ResponseWriter, r *http.Request) {
		identity, ok := authcontext.IdentityFromContext(r.Context())
		if !ok {
			httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
			return
		}
		status, err := bots.Get(r.Context(), identity.UserID)
		if err != nil {
			httpjson.WriteError(w, troncerrors.Internal("failed to load the chat agent", err))
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, status)
	})
}
