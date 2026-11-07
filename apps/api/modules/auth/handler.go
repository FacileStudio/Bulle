package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

type PlanFunc func(ctx context.Context, userID int64) string

type Handler struct {
	service *Service
	store   store.Store
	plan    PlanFunc
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, errors.Unauthorized("not authenticated"))
		return
	}
	user, err := h.service.UserByID(r.Context(), identity.UserID)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	var assistant *AssistantProfileResponse
	if h.store != nil {
		if profile, err := h.store.GetAssistantProfile(r.Context(), identity.UserID); err == nil {
			traits := profile.Traits
			if traits == nil {
				traits = []string{}
			}
			assistant = &AssistantProfileResponse{
				UserID:    profile.UserID,
				Name:      profile.Name,
				AvatarURL: profile.AvatarURL,
				Archetype: profile.Archetype,
				Traits:    traits,
			}
		}
	}
	var plan string
	if h.plan != nil {
		plan = h.plan(r.Context(), identity.UserID)
	}
	httpjson.WriteJSON(w, http.StatusOK, MeResponse{User: UserResponse{
		Plan:      plan,
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		AvatarURL: user.PictureURL,
		Locale:    user.Locale,
		CreatedAt: user.CreatedAt.UTC().Format(time.RFC3339),
		Assistant: assistant,
	}})
}
