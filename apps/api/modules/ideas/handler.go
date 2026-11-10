package ideas

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

// ManualBatchesPerDay caps the "Think of more" batches of one space per local day.
const ManualBatchesPerDay = 3

// AppBatchesPerDay caps the batches one app asks for a user per local day,
// across all the user's spaces.
const AppBatchesPerDay = 3

var openStatuses = []string{store.IdeaNew, store.IdeaSaved, store.IdeaDismissed}

type Handler struct {
	store  store.Store
	runner *runner.Runner
	logger *slog.Logger
	now    func() time.Time
}

func NewHandler(st store.Store, rn *runner.Runner, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{store: st, runner: rn, logger: logger, now: time.Now}
}

func userOf(w http.ResponseWriter, r *http.Request) (int64, bool) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
	}
	return identity.UserID, ok
}

func storeError(what string, err error) error {
	if errors.Is(err, store.ErrNotFound) {
		return troncerrors.NotFound(what + " not found")
	}
	return troncerrors.Internal("failed to load "+what, err)
}

func (h *Handler) settings(ctx context.Context, userID int64, spaceID string) (store.IdeaSettings, error) {
	settings, err := h.store.GetIdeaSettings(ctx, userID, spaceID)
	if errors.Is(err, store.ErrNotFound) {
		return h.store.SaveIdeaSettings(ctx, store.DefaultIdeaSettings(userID, spaceID))
	}
	return settings, err
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := userOf(w, r)
	if !ok {
		return
	}
	app, err := appParam(r)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	res, err := h.page(r.Context(), userID, r.URL.Query().Get("space_id"), app)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) page(ctx context.Context, userID int64, spaceID, app string) (ListResponse, error) {
	space, err := h.space(ctx, userID, spaceID)
	if err != nil {
		return ListResponse{}, storeError("space", err)
	}
	settings, err := h.settings(ctx, userID, space.ID)
	if err != nil {
		return ListResponse{}, storeError("settings", err)
	}
	ideas, err := h.store.ListActiveIdeas(ctx, userID, space.ID, app, h.now())
	if err != nil {
		return ListResponse{}, storeError("ideas", err)
	}
	goals, err := h.store.ListGoals(ctx, userID, space.ID)
	if err != nil {
		return ListResponse{}, storeError("goals", err)
	}
	_, generating, err := h.store.RunningIdeaBatch(ctx, userID, space.ID, batchApp(app))
	if err != nil {
		return ListResponse{}, storeError("batch", err)
	}
	res := ListResponse{
		Ideas: make([]IdeaResponse, 0, len(ideas)), Goals: make([]GoalResponse, 0, len(goals)),
		Settings: settingsResponse(settings), Generating: generating,
	}
	for _, idea := range ideas {
		res.Ideas = append(res.Ideas, ideaResponse(idea))
	}
	for _, goal := range goals {
		res.Goals = append(res.Goals, goalResponse(goal))
	}
	if last, ok, err := h.store.LastIdeaBatch(ctx, userID, space.ID, batchApp(app)); err == nil && ok {
		at := last.CreatedAt
		if last.FinishedAt != nil {
			at = *last.FinishedAt
		}
		res.LastBatchAt = &at
	}
	return res, nil
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	userID, ok := userOf(w, r)
	if !ok {
		return
	}
	var body UpdateIdeaRequest
	if err := httpjson.DecodeJSON(w, r, &body); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	idea, err := h.store.GetIdea(r.Context(), userID, chi.URLParam(r, "id"))
	if err != nil {
		httpjson.WriteError(w, storeError("idea", err))
		return
	}
	if err := applyIdeaUpdate(&idea, body); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	updated, err := h.store.UpdateIdea(r.Context(), idea)
	if err != nil {
		httpjson.WriteError(w, storeError("idea", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, ideaResponse(updated))
}

func applyIdeaUpdate(idea *store.Idea, body UpdateIdeaRequest) error {
	if body.Status != nil {
		if !slices.Contains(openStatuses, *body.Status) {
			return troncerrors.Invalid("status must be saved, new or dismissed")
		}
		if !slices.Contains(openStatuses, idea.Status) {
			return troncerrors.Conflict("the idea is no longer open")
		}
		idea.Status = *body.Status
	}
	if body.Feedback != nil {
		if !slices.Contains([]string{"", store.FeedbackUp, store.FeedbackDown}, *body.Feedback) {
			return troncerrors.Invalid("feedback must be empty, up or down")
		}
		idea.Feedback = *body.Feedback
	}
	if body.Reason != nil {
		reason := strings.TrimSpace(*body.Reason)
		if len([]rune(reason)) > store.IdeaReasonLimit {
			return troncerrors.Invalid("reason is too long")
		}
		idea.Reason = reason
	}
	return nil
}
