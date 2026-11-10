package artifacts

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

const (
	maxTitle   = 200
	maxContent = 1 << 20
)

// Handler proxies the user's Mycelium artifacts, across every tree they can
// read, with their own delegated token.
type Handler struct {
	store   store.Store
	gateway *suite.Gateway
}

func (h *Handler) listArtifacts(w http.ResponseWriter, r *http.Request) {
	userID, err := h.caller(r)
	if err == nil {
		var items []suite.ArtifactSummary
		items, err = suite.NewMycelium(h.gateway).ListAllArtifacts(r.Context(), userID)
		if err == nil {
			httpjson.WriteJSON(w, http.StatusOK, ListResponse{Artifacts: toArtifacts(items), Available: true})
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
	userID, where, err := h.locate(r)
	if err != nil {
		httpjson.WriteError(w, routeError(err))
		return
	}
	found, err := suite.NewMycelium(h.gateway).GetArtifact(r.Context(), userID, where, chi.URLParam(r, "id"))
	if err != nil {
		httpjson.WriteError(w, routeError(err))
		return
	}
	found.SpaceID = where
	httpjson.WriteJSON(w, http.StatusOK, DetailResponse{Artifact: toArtifact(found.ArtifactSummary), Content: found.Content})
}

func (h *Handler) updateArtifact(w http.ResponseWriter, r *http.Request) {
	userID, where, err := h.locate(r)
	if err != nil {
		httpjson.WriteError(w, routeError(err))
		return
	}
	body, err := readUpdate(w, r)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	if err := suite.ValidateArtifact(r.Context(), body.Content); err != nil {
		httpjson.WriteError(w, troncerrors.Invalid(err.Error()))
		return
	}
	saved, err := suite.NewMycelium(h.gateway).UpdateArtifact(r.Context(), userID, where, chi.URLParam(r, "id"), suite.ArtifactWrite{Title: body.Title, Content: body.Content})
	if err != nil {
		httpjson.WriteError(w, routeError(err))
		return
	}
	saved.SpaceID = where
	httpjson.WriteJSON(w, http.StatusOK, DetailResponse{Artifact: toArtifact(saved.ArtifactSummary), Content: saved.Content})
}

// readUpdate decodes an update and refuses what Mycelium would refuse, with
// the same limits the agent tools apply.
func readUpdate(w http.ResponseWriter, r *http.Request) (UpdateRequest, error) {
	var body UpdateRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxContent*2)).Decode(&body); err != nil {
		return body, troncerrors.Invalid("invalid body")
	}
	body.Title = strings.TrimSpace(body.Title)
	switch {
	case utf8.RuneCountInString(body.Title) > maxTitle:
		return body, troncerrors.Invalid("title must be at most 200 characters")
	case strings.TrimSpace(body.Content) == "":
		return body, troncerrors.Invalid("content is required")
	case len(body.Content) > maxContent:
		return body, troncerrors.TooLarge("content exceeds 1 MB")
	}
	return body, nil
}

func (h *Handler) deleteArtifact(w http.ResponseWriter, r *http.Request) {
	userID, where, err := h.locate(r)
	if err == nil {
		err = suite.NewMycelium(h.gateway).DeleteArtifact(r.Context(), userID, where, chi.URLParam(r, "id"))
	}
	if err != nil {
		httpjson.WriteError(w, routeError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
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

func toArtifacts(items []suite.ArtifactSummary) []Artifact {
	out := make([]Artifact, 0, len(items))
	for _, item := range items {
		out = append(out, toArtifact(item))
	}
	return out
}

func toArtifact(item suite.ArtifactSummary) Artifact {
	return Artifact{
		ID: item.ID, Title: item.Title, Machine: item.Machine, Created: item.Created, Updated: item.Updated,
		MyceliumSpace: item.SpaceID, SpaceName: item.SpaceName,
	}
}
