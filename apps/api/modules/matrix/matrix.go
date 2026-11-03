// Package matrix serves the settings of the user's Matrix bot: the bot
// account, who may call it, and which room each space is linked to.
package matrix

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/matrixbot"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

const maxBody = 64 << 10

// Bots is the part of the bot manager these routes use.
type Bots interface {
	Get(ctx context.Context, userID int64) (matrixbot.Settings, error)
	Save(ctx context.Context, userID int64, in matrixbot.Settings) (matrixbot.Settings, error)
}

// RegisterRoutes mounts GET and PUT /matrix.
func RegisterRoutes(router chi.Router, bots Bots, requireAuth func(http.Handler) http.Handler) {
	h := handler{bots: bots}
	router.With(requireAuth).Get("/matrix", h.get)
	router.With(requireAuth).Put("/matrix", h.put)
}

type handler struct{ bots Bots }

func (h handler) get(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	settings, err := h.bots.Get(r.Context(), identity.UserID)
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to load matrix settings", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, settings)
}

func (h handler) put(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	var in matrixbot.Settings
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBody)).Decode(&in); err != nil {
		httpjson.WriteError(w, troncerrors.Invalid("malformed matrix settings"))
		return
	}
	saved, err := h.bots.Save(r.Context(), identity.UserID, in)
	switch {
	case errors.Is(err, matrixbot.ErrDisabled):
		httpjson.WriteError(w, troncerrors.Unavailable(err.Error()))
	case errors.Is(err, matrixbot.ErrInvalid), errors.Is(err, matrixbot.ErrUnknownSpace):
		httpjson.WriteError(w, troncerrors.Invalid(err.Error()))
	case err != nil:
		httpjson.WriteError(w, troncerrors.Internal("failed to save matrix settings", err))
	default:
		httpjson.WriteJSON(w, http.StatusOK, saved)
	}
}
