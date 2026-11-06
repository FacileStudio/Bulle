package artifacts

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

// Handler proxies a Bulle space's artifacts to its Mycelium space with the
// user's own delegated token.
type Handler struct {
	store   store.Store
	gateway *suite.Gateway
}

func (h *Handler) listArtifacts(w http.ResponseWriter, r *http.Request) {
	userID, space, err := h.resolve(r)
	if err == nil {
		var items []suite.ArtifactSummary
		items, err = suite.NewMycelium(h.gateway).ListArtifacts(r.Context(), userID, space)
		if err == nil {
			httpjson.WriteJSON(w, http.StatusOK, ListResponse{Artifacts: markdownOnly(items), Available: true})
			return
		}
	}
	if disconnected(err) {
		httpjson.WriteJSON(w, http.StatusOK, ListResponse{Artifacts: []Artifact{}, Available: false})
		return
	}
	httpjson.WriteError(w, routeError(err))
}

func (h *Handler) getArtifact(w http.ResponseWriter, r *http.Request) {
	userID, space, err := h.resolve(r)
	if err != nil {
		httpjson.WriteError(w, routeError(err))
		return
	}
	found, err := suite.NewMycelium(h.gateway).GetArtifact(r.Context(), userID, space, chi.URLParam(r, "id"))
	if err == nil && found.Format == "html" {
		err = &suite.RefusedError{App: suite.AppMycelium, Status: http.StatusNotFound}
	}
	if err != nil {
		httpjson.WriteError(w, routeError(err))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, DetailResponse{Artifact: toArtifact(found.ArtifactSummary), Content: found.Content})
}

func (h *Handler) deleteArtifact(w http.ResponseWriter, r *http.Request) {
	userID, space, err := h.resolve(r)
	if err == nil {
		err = suite.NewMycelium(h.gateway).DeleteArtifact(r.Context(), userID, space, chi.URLParam(r, "id"))
	}
	if err != nil {
		httpjson.WriteError(w, routeError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// resolve answers the caller and the Mycelium space of the Bulle space named
// by space_id, or of the default space when it is omitted.
func (h *Handler) resolve(r *http.Request) (int64, string, error) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		return 0, "", troncerrors.Unauthorized("not authenticated")
	}
	if h.gateway == nil || !h.gateway.Config().Apps[suite.AppMycelium].Enabled() {
		return identity.UserID, "", errNotConnected
	}
	space, err := h.space(r.Context(), identity.UserID, strings.TrimSpace(r.URL.Query().Get("space_id")))
	if err != nil {
		return identity.UserID, "", err
	}
	id, err := suite.NewMycelium(h.gateway).SpaceOf(r.Context(), identity.UserID, space, h.store)
	return identity.UserID, id, err
}

func (h *Handler) space(ctx context.Context, userID int64, id string) (store.Space, error) {
	if id == "" {
		return h.store.EnsureDefaultSpace(ctx, userID)
	}
	space, err := h.store.GetSpace(ctx, userID, id)
	if errors.Is(err, store.ErrNotFound) {
		return space, troncerrors.NotFound("space not found")
	}
	return space, err
}

func routeError(err error) error {
	var known *troncerrors.Error
	if errors.As(err, &known) {
		return known
	}
	return appError(err)
}

func markdownOnly(items []suite.ArtifactSummary) []Artifact {
	out := make([]Artifact, 0, len(items))
	for _, item := range items {
		if item.Format != "html" {
			out = append(out, toArtifact(item))
		}
	}
	return out
}

func toArtifact(item suite.ArtifactSummary) Artifact {
	return Artifact{ID: item.ID, Title: item.Title, Machine: item.Machine, Created: item.Created, Updated: item.Updated}
}
