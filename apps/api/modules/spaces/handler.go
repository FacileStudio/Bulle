package spaces

import (
	"errors"
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	store store.Store
}

func NewHandler(st store.Store) *Handler {
	return &Handler{store: st}
}

func (h *Handler) listSpaces(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	items, err := h.store.ListSpaces(r.Context(), identity.UserID)
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to list spaces", err))
		return
	}
	res := make([]SpaceResponse, 0, len(items))
	for _, item := range items {
		res = append(res, toResponse(item))
	}
	httpjson.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) createSpace(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	var body SpaceRequest
	if err := httpjson.DecodeJSON(w, r, &body); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	space, err := apply(store.Space{UserID: identity.UserID}, body)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	created, err := h.store.CreateSpace(r.Context(), space)
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to create space", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusCreated, toResponse(created))
}

func (h *Handler) updateSpace(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	var body SpaceRequest
	if err := httpjson.DecodeJSON(w, r, &body); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	existing, err := h.store.GetSpace(r.Context(), identity.UserID, chi.URLParam(r, "id"))
	if err != nil {
		httpjson.WriteError(w, storeError(err, "failed to load space"))
		return
	}
	space, err := apply(existing, body)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	updated, err := h.store.UpdateSpace(r.Context(), space)
	if err != nil {
		httpjson.WriteError(w, storeError(err, "failed to update space"))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, toResponse(updated))
}

func (h *Handler) deleteSpace(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	if err := h.store.DeleteSpace(r.Context(), identity.UserID, chi.URLParam(r, "id")); err != nil {
		httpjson.WriteError(w, storeError(err, "failed to delete space"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func storeError(err error, message string) error {
	switch {
	case errors.Is(err, store.ErrNotFound):
		return troncerrors.NotFound("space not found")
	case errors.Is(err, store.ErrLastSpace):
		return troncerrors.Conflict("the last space cannot be deleted")
	}
	return troncerrors.Internal(message, err)
}
