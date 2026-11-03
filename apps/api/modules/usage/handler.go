package usage

import (
	"net/http"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

// Handler serves the usage summary.
type Handler struct {
	store        store.Store
	defaultModel string
	now          func() time.Time
}

// NewHandler builds a usage handler. defaultModel names the model of sessions
// that never picked one.
func NewHandler(st store.Store, defaultModel string) *Handler {
	return &Handler{store: st, defaultModel: defaultModel, now: time.Now}
}

func (h *Handler) summary(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	location := time.UTC
	if name := r.URL.Query().Get("tz"); name != "" {
		loaded, err := time.LoadLocation(name)
		if err != nil {
			httpjson.WriteError(w, troncerrors.Invalid("tz is not an IANA timezone"))
			return
		}
		location = loaded
	}
	now := h.now().In(location)
	runs, err := h.store.UsageRuns(r.Context(), identity.UserID, monthStart(now))
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to load usage", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, summarize(runs, now, h.defaultModel))
}
