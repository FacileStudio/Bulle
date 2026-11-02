package assistant

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

type Handler struct {
	store store.Store
}

func NewHandler(st store.Store) *Handler {
	return &Handler{store: st}
}

func (h *Handler) getProfile(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	profile, err := h.store.GetAssistantProfile(r.Context(), identity.UserID)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		httpjson.WriteError(w, troncerrors.Internal("failed to load assistant profile", err))
		return
	}
	traits := profile.Traits
	if traits == nil {
		traits = []string{}
	}
	httpjson.WriteJSON(w, http.StatusOK, ProfileResponse{
		UserID:    identity.UserID,
		Name:      profile.Name,
		AvatarURL: profile.AvatarURL,
		Archetype: profile.Archetype,
		Traits:    traits,
	})
}

func (h *Handler) saveProfile(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpjson.WriteError(w, troncerrors.Invalid("invalid json payload"))
		return
	}
	traits := req.Traits
	if traits == nil {
		traits = []string{}
	}
	profile := store.AssistantProfile{
		UserID:    identity.UserID,
		Name:      req.Name,
		AvatarURL: req.AvatarURL,
		Archetype: req.Archetype,
		Traits:    traits,
	}
	if err := h.store.SaveAssistantProfile(r.Context(), profile); err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to save assistant profile", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, ProfileResponse{
		UserID:    identity.UserID,
		Name:      profile.Name,
		AvatarURL: profile.AvatarURL,
		Archetype: profile.Archetype,
		Traits:    traits,
	})
}
