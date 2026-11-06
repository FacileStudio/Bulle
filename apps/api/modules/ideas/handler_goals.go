package ideas

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) createGoal(w http.ResponseWriter, r *http.Request) {
	userID, ok := userOf(w, r)
	if !ok {
		return
	}
	var body CreateGoalRequest
	if err := httpjson.DecodeJSON(w, r, &body); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	space, err := h.space(r.Context(), userID, body.SpaceID)
	if err != nil {
		httpjson.WriteError(w, storeError("space", err))
		return
	}
	goal := store.Goal{UserID: userID, SpaceID: space.ID}
	err = applyGoal(&goal, UpdateGoalRequest{Title: &body.Title, Category: &body.Category})
	if err == nil {
		err = h.roomForOpenGoal(r.Context(), userID, space.ID)
	}
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	created, err := h.store.CreateGoal(r.Context(), goal)
	if err != nil {
		httpjson.WriteError(w, storeError("space", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusCreated, goalResponse(created))
}

func (h *Handler) updateGoal(w http.ResponseWriter, r *http.Request) {
	userID, ok := userOf(w, r)
	if !ok {
		return
	}
	var body UpdateGoalRequest
	if err := httpjson.DecodeJSON(w, r, &body); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	goal, err := h.store.GetGoal(r.Context(), userID, chi.URLParam(r, "id"))
	if err != nil {
		httpjson.WriteError(w, storeError("goal", err))
		return
	}
	reopens := goal.Done && body.Done != nil && !*body.Done
	err = applyGoal(&goal, body)
	if err == nil && reopens {
		err = h.roomForOpenGoal(r.Context(), userID, goal.SpaceID)
	}
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	updated, err := h.store.UpdateGoal(r.Context(), goal)
	if err != nil {
		httpjson.WriteError(w, storeError("goal", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, goalResponse(updated))
}

func (h *Handler) deleteGoal(w http.ResponseWriter, r *http.Request) {
	userID, ok := userOf(w, r)
	if !ok {
		return
	}
	if err := h.store.DeleteGoal(r.Context(), userID, chi.URLParam(r, "id")); err != nil {
		httpjson.WriteError(w, storeError("goal", err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func applyGoal(goal *store.Goal, body UpdateGoalRequest) error {
	if body.Title != nil {
		title := strings.TrimSpace(*body.Title)
		if title == "" || len([]rune(title)) > store.GoalTitleLimit {
			return troncerrors.Invalid("goal title must be 1 to 120 characters")
		}
		goal.Title = title
	}
	if body.Category != nil {
		if !slices.Contains(store.GoalCategories(), *body.Category) {
			return troncerrors.Invalid("category must be one of " + strings.Join(store.GoalCategories(), ", "))
		}
		goal.Category = *body.Category
	}
	if body.Done != nil {
		goal.Done = *body.Done
	}
	return nil
}

func (h *Handler) roomForOpenGoal(ctx context.Context, userID int64, spaceID string) error {
	goals, err := h.store.ListGoals(ctx, userID, spaceID)
	if err != nil {
		return storeError("goals", err)
	}
	open := 0
	for _, goal := range goals {
		if !goal.Done {
			open++
		}
	}
	if open >= store.OpenGoalsLimit {
		return troncerrors.Invalid("a space holds at most 20 open goals")
	}
	return nil
}
