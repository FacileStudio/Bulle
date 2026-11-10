package ideas

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

// space resolves the space a request names: a Bulle space id or the id of
// the Registre shared space it mirrors, always one of the user's own, and the
// default space when none is named.
func (h *Handler) space(ctx context.Context, userID int64, spaceID string) (store.Space, error) {
	spaceID = strings.TrimSpace(spaceID)
	if spaceID == "" {
		return h.store.EnsureDefaultSpace(ctx, userID)
	}
	space, err := h.store.GetSpace(ctx, userID, spaceID)
	if !errors.Is(err, store.ErrNotFound) {
		return space, err
	}
	spaces, err := h.store.ListSpaces(ctx, userID)
	if err != nil {
		return store.Space{}, err
	}
	for _, space := range spaces {
		if space.RegistreSpaceID == spaceID {
			return space, nil
		}
	}
	return store.Space{}, store.ErrNotFound
}

// appParam reads the optional ?app= filter: one of the suite apps or bulle.
func appParam(r *http.Request) (string, error) {
	app := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("app")))
	if app != "" && !slices.Contains(store.IdeaApps(), app) {
		return "", troncerrors.Invalid("app must be one of " + strings.Join(store.IdeaApps(), ", "))
	}
	return app, nil
}

// batchApp is the batch scope that fills an app's list: Bulle's own list is
// filled by the batches over every app.
func batchApp(app string) string {
	if app == store.IdeaAppBulle {
		return ""
	}
	return app
}

func (h *Handler) run(w http.ResponseWriter, r *http.Request) {
	userID, ok := userOf(w, r)
	if !ok {
		return
	}
	idea, err := h.ideaInSpace(r.Context(), userID, chi.URLParam(r, "id"), r.URL.Query().Get("space_id"))
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	if idea.Status == store.IdeaDone && idea.SessionID != "" {
		if latest, ok, err := h.store.LatestRun(r.Context(), idea.SessionID); err == nil && ok {
			httpjson.WriteJSON(w, http.StatusOK, RunStartResponse{RunID: latest.ID, SessionID: idea.SessionID, IdeaID: idea.ID})
			return
		}
	}
	sessionID, runID, err := h.startRun(r.Context(), idea)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	idea.Status, idea.SessionID = store.IdeaDone, sessionID
	if _, err := h.store.UpdateIdea(r.Context(), idea); err != nil {
		httpjson.WriteError(w, storeError("idea", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusAccepted, RunStartResponse{RunID: runID, SessionID: sessionID, IdeaID: idea.ID})
}

// ideaInSpace loads the user's idea and, when the caller names its active
// space, refuses an idea of another space.
func (h *Handler) ideaInSpace(ctx context.Context, userID int64, id, spaceID string) (store.Idea, error) {
	idea, err := h.store.GetIdea(ctx, userID, id)
	if err != nil {
		return store.Idea{}, storeError("idea", err)
	}
	if strings.TrimSpace(spaceID) == "" {
		return idea, nil
	}
	space, err := h.space(ctx, userID, spaceID)
	if err != nil {
		return store.Idea{}, storeError("space", err)
	}
	if space.ID != idea.SpaceID {
		return store.Idea{}, troncerrors.NotFound("idea not found")
	}
	return idea, nil
}

func (h *Handler) getRun(w http.ResponseWriter, r *http.Request) {
	userID, ok := userOf(w, r)
	if !ok {
		return
	}
	view, err := h.runner.ViewRun(r.Context(), userID, chi.URLParam(r, "id"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			httpjson.WriteError(w, troncerrors.NotFound("run not found"))
			return
		}
		httpjson.WriteError(w, troncerrors.Internal("failed to load the run", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, runResponse(view))
}
