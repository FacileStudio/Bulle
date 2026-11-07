package ideas

import (
	"context"
	"errors"
	"github.com/FacileStudio/Bulle/apps/api/internal/plans"
	"net/http"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

func runError(err error) error {
	switch {
	case errors.Is(err, plans.ErrAIAccess):
		return err
	case errors.Is(err, runner.ErrIdeasBusy):
		return troncerrors.Conflict("a batch is already running")
	case errors.Is(err, runner.ErrTooManyRuns):
		return troncerrors.RateLimited("too many active runs")
	case errors.Is(err, runner.ErrShuttingDown):
		return troncerrors.Unavailable("server is shutting down")
	case errors.Is(err, store.ErrNotFound):
		return troncerrors.NotFound("space not found")
	}
	return troncerrors.Internal("failed to start the run", err)
}

func (h *Handler) start(w http.ResponseWriter, r *http.Request) {
	userID, ok := userOf(w, r)
	if !ok {
		return
	}
	idea, err := h.store.GetIdea(r.Context(), userID, chi.URLParam(r, "id"))
	if err != nil {
		httpjson.WriteError(w, storeError("idea", err))
		return
	}
	if idea.Status == store.IdeaDone && idea.SessionID != "" {
		httpjson.WriteJSON(w, http.StatusOK, StartResponse{SessionID: idea.SessionID})
		return
	}
	sessionID, err := h.startChat(r.Context(), idea)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	idea.Status, idea.SessionID = store.IdeaDone, sessionID
	if _, err := h.store.UpdateIdea(r.Context(), idea); err != nil {
		httpjson.WriteError(w, storeError("idea", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, StartResponse{SessionID: sessionID})
}

func (h *Handler) startChat(ctx context.Context, idea store.Idea) (string, error) {
	session, err := h.runner.CreateWebSession(ctx, idea.UserID, idea.SpaceID, idea.Title)
	if err != nil {
		return "", runError(err)
	}
	text := idea.Prompt
	if idea.Draft != "" {
		text += "\n\nDraft:\n" + idea.Draft
	}
	if _, err := h.runner.Start(ctx, idea.UserID, session.ID, text); err != nil {
		_ = h.store.DeleteSession(context.WithoutCancel(ctx), idea.UserID, session.ID)
		return "", runError(err)
	}
	return session.ID, nil
}

func (h *Handler) generate(w http.ResponseWriter, r *http.Request) {
	userID, ok := userOf(w, r)
	if !ok {
		return
	}
	var body GenerateRequest
	if err := httpjson.DecodeJSON(w, r, &body); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	batchID, err := h.manualBatch(r.Context(), userID, body.SpaceID)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusAccepted, GenerateResponse{BatchID: batchID})
}

func (h *Handler) manualBatch(ctx context.Context, userID int64, spaceID string) (string, error) {
	space, err := h.space(ctx, userID, spaceID)
	if err != nil {
		return "", storeError("space", err)
	}
	settings, err := h.settings(ctx, userID, space.ID)
	if err != nil {
		return "", storeError("settings", err)
	}
	count, err := h.store.CountIdeaBatches(ctx, userID, space.ID, store.BatchManual, localMidnight(h.now(), settings.Timezone))
	if err != nil {
		return "", storeError("batches", err)
	}
	if count >= ManualBatchesPerDay {
		return "", troncerrors.RateLimited("at most 3 manual batches per day; more ideas come tomorrow morning")
	}
	batchID, err := h.runner.GenerateIdeas(ctx, userID, space.ID, store.BatchManual)
	if err != nil {
		return "", runError(err)
	}
	return batchID, nil
}

func localMidnight(now time.Time, timezone string) time.Time {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	local := now.In(loc)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
}
