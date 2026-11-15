package usage

import (
	"context"
	"net/http"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/aiusage"
	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

// Budget reads where a user stands against their monthly AI budget. A nil
// Budget leaves the summary without one.
type Budget interface {
	Status(ctx context.Context, userID int64) (aiusage.Status, error)
}

// Handler serves the usage summary.
type Handler struct {
	store        store.Store
	budget       Budget
	defaultModel string
	now          func() time.Time
}

// NewHandler builds a usage handler. defaultModel names the model of sessions
// that never picked one.
func NewHandler(st store.Store, defaultModel string, budget Budget) *Handler {
	return &Handler{store: st, budget: budget, defaultModel: defaultModel, now: time.Now}
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
	out := summarize(runs, now, h.defaultModel)
	if h.budget != nil {
		status, err := h.budget.Status(r.Context(), identity.UserID)
		if err != nil {
			httpjson.WriteError(w, troncerrors.Internal("failed to load the AI allowance", err))
			return
		}
		out.Budget = &status
	}
	httpjson.WriteJSON(w, http.StatusOK, out)
}
