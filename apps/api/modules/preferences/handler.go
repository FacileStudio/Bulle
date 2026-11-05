package preferences

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/core/settings"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

const maxConfigBytes = 64 << 10

type Handler struct {
	store        store.Store
	models       []string
	defaultModel string
}

func NewHandler(st store.Store, models []string, defaultModel string) *Handler {
	return &Handler{store: st, models: models, defaultModel: defaultModel}
}

func (h *Handler) userID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return 0, false
	}
	return identity.UserID, true
}

// current reads the saved preferences, treating "never saved" as "everything
// on the instance default" rather than as an error.
func (h *Handler) current(r *http.Request, userID int64) store.Preferences {
	prefs, err := h.store.GetPreferences(r.Context(), userID)
	if err != nil {
		return store.Preferences{UserID: userID}
	}
	return prefs
}

func (h *Handler) respond(w http.ResponseWriter, prefs store.Preferences) {
	models := h.models
	if models == nil {
		models = []string{}
	}
	httpjson.WriteJSON(w, http.StatusOK, PreferencesResponse{
		Model:        prefs.Model,
		CLIConfig:    prefs.CLIConfig,
		Models:       models,
		DefaultModel: h.defaultModel,
	})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(w, r)
	if !ok {
		return
	}
	h.respond(w, h.current(r, userID))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(w, r)
	if !ok {
		return
	}
	var body UpdatePreferencesRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxConfigBytes)).Decode(&body); err != nil {
		httpjson.WriteError(w, troncerrors.Invalid("invalid json payload"))
		return
	}
	prefs := h.current(r, userID)
	if body.Model != nil {
		model := strings.TrimSpace(*body.Model)
		if err := h.validateModel(model); err != nil {
			httpjson.WriteError(w, troncerrors.Invalid(err.Error()))
			return
		}
		prefs.Model = model
	}
	if body.CLIConfig != nil {
		if err := validateConfig(*body.CLIConfig); err != nil {
			httpjson.WriteError(w, troncerrors.Invalid(err.Error()))
			return
		}
		prefs.CLIConfig = *body.CLIConfig
	}
	if err := h.store.SavePreferences(r.Context(), prefs); err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to save preferences", err))
		return
	}
	h.respond(w, prefs)
}

// CLIConfig is what the CLI asks for on every start. It is a separate route
// because the client that reads it holds a login token, not a browser session,
// and needs the YAML alone.
func (h *Handler) CLIConfig(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(w, r)
	if !ok {
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, CLIConfigResponse{Config: h.current(r, userID).CLIConfig})
}

func (h *Handler) validateModel(model string) error {
	if model == "" || slices.Contains(h.models, model) {
		return nil
	}
	return fmt.Errorf("model %q is not offered; this server offers %s", model, strings.Join(h.models, ", "))
}

// validateConfig refuses a document the CLI could not load, and the one key
// the server's copy must not carry: server.url. The CLI sends its login token
// to the address in its own config, so a saved address would hand the token to
// whoever wrote it.
func validateConfig(raw string) error {
	parsed, err := settings.Parse([]byte(raw))
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	if parsed.Server.URL != "" {
		return errors.New("server.url belongs to the machine, not to the server's copy of the config")
	}
	return nil
}
