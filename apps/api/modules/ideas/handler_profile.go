package ideas

import (
	"errors"
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

// assistantProfile answers the name and avatar of the user's assistant. It
// spends no AI, so a user without ai.access still gets their blob.
func (h *Handler) assistantProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := userOf(w, r)
	if !ok {
		return
	}
	profile, err := h.store.GetAssistantProfile(r.Context(), userID)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		httpjson.WriteError(w, troncerrors.Internal("failed to load assistant profile", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, AssistantProfileResponse{Name: profile.Name, AvatarURL: profile.AvatarURL})
}
