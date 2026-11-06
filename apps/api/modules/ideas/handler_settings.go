package ideas

import (
	"net/http"
	"strings"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

func (h *Handler) saveSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := userOf(w, r)
	if !ok {
		return
	}
	var body SettingsRequest
	if err := httpjson.DecodeJSON(w, r, &body); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	space, err := h.space(r.Context(), userID, body.SpaceID)
	if err != nil {
		httpjson.WriteError(w, storeError("space", err))
		return
	}
	settings, err := h.settings(r.Context(), userID, space.ID)
	if err != nil {
		httpjson.WriteError(w, storeError("settings", err))
		return
	}
	if err := applySettings(&settings, body); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	saved, err := h.store.SaveIdeaSettings(r.Context(), settings)
	if err != nil {
		httpjson.WriteError(w, storeError("settings", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, settingsResponse(saved))
}

func applySettings(settings *store.IdeaSettings, body SettingsRequest) error {
	if body.Hour != nil {
		if *body.Hour < 0 || *body.Hour > 23 {
			return troncerrors.Invalid("hour must be between 0 and 23")
		}
		settings.Hour = *body.Hour
	}
	if body.Timezone != nil {
		zone := strings.TrimSpace(*body.Timezone)
		if _, err := time.LoadLocation(zone); err != nil || zone == "" || zone == "Local" {
			return troncerrors.Invalid("timezone must be an IANA time zone name")
		}
		settings.Timezone = zone
	}
	if body.Steering != nil {
		steering := strings.TrimSpace(*body.Steering)
		if len([]rune(steering)) > store.SteeringLimit {
			return troncerrors.Invalid("steering is too long")
		}
		settings.Steering = steering
	}
	if body.Enabled != nil {
		settings.Enabled = *body.Enabled
	}
	if body.Push != nil {
		settings.Push = *body.Push
	}
	return nil
}
