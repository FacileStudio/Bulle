package preferences

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/catalog"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

// Catalog is the shared model catalogue, bound to the instance's backend. A nil
// one means only the instance default can be chosen.
type Catalog func(ctx context.Context, userID int64) []catalog.Model

type Handler struct {
	store        store.Store
	catalog      Catalog
	defaultModel string
}

func NewHandler(st store.Store, models Catalog, defaultModel string) *Handler {
	return &Handler{store: st, catalog: models, defaultModel: defaultModel}
}

func (h *Handler) userID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return 0, false
	}
	return identity.UserID, true
}

// current reads the saved preference, treating "never saved" as "the instance
// default" rather than as an error.
func (h *Handler) current(r *http.Request, userID int64) string {
	prefs, err := h.store.GetPreferences(r.Context(), userID)
	if err != nil {
		return ""
	}
	return prefs.Model
}

func (h *Handler) respond(w http.ResponseWriter, r *http.Request, userID int64, model string) {
	options := []catalog.Model{}
	if h.catalog != nil {
		options = h.catalog(r.Context(), userID)
	}
	httpjson.WriteJSON(w, http.StatusOK, PreferencesResponse{
		Model:        model,
		Catalog:      options,
		DefaultModel: h.defaultModel,
	})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(w, r)
	if !ok {
		return
	}
	h.respond(w, r, userID, h.current(r, userID))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(w, r)
	if !ok {
		return
	}
	var body UpdatePreferencesRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&body); err != nil {
		httpjson.WriteError(w, troncerrors.Invalid("invalid json payload"))
		return
	}
	model := h.current(r, userID)
	if body.Model != nil {
		candidate := strings.TrimSpace(*body.Model)
		if err := h.validate(r, userID, candidate); err != nil {
			httpjson.WriteError(w, troncerrors.Invalid(err.Error()))
			return
		}
		if err := h.store.SavePreferences(r.Context(), store.Preferences{UserID: userID, Model: candidate}); err != nil {
			httpjson.WriteError(w, troncerrors.Internal("failed to save preferences", err))
			return
		}
		model = candidate
	}
	h.respond(w, r, userID, model)
}

// validate accepts the instance default and anything the backend itself
// advertises. An empty choice means the instance default.
func (h *Handler) validate(r *http.Request, userID int64, model string) error {
	if model == "" || model == h.defaultModel {
		return nil
	}
	if h.catalog == nil {
		return &unknownModel{model: model}
	}
	for _, option := range h.catalog(r.Context(), userID) {
		if option.ID == model {
			return nil
		}
	}
	return &unknownModel{model: model}
}

type unknownModel struct{ model string }

func (e *unknownModel) Error() string {
	return "model " + e.model + " is not offered by this server"
}
