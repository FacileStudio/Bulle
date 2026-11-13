package sessions

import (
	"errors"
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

// markRead records that the caller has read the session up to now, which
// clears its unread mark on every device.
func (h *Handler) markRead(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	if err := h.store.MarkSessionRead(r.Context(), identity.UserID, chi.URLParam(r, "id")); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			httpjson.WriteError(w, troncerrors.NotFound("session not found"))
			return
		}
		httpjson.WriteError(w, troncerrors.Internal("failed to mark the session read", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, EmptyResponse{})
}
