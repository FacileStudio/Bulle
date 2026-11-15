package artifacts

import (
	"errors"
	"net/http"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/shares"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

const sharePath = "/share/"

// Sharing is what the public links need: their store and the origin the
// public URL is built on. An empty Origin falls back to the request's host.
type Sharing struct {
	Store  *shares.Store
	Origin string
}

func (h *Handler) shareOf(share shares.Share, r *http.Request) ShareResponse {
	origin := h.sharing.Origin
	if origin == "" {
		scheme := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		origin = scheme + "://" + r.Host
	}
	return ShareResponse{Token: share.Token, URL: origin + sharePath + share.Token, CreatedAt: share.CreatedAt}
}

func (h *Handler) createShare(w http.ResponseWriter, r *http.Request) {
	userID, where, err := h.locate(r)
	if err != nil {
		httpjson.WriteError(w, routeError(err))
		return
	}
	id := chi.URLParam(r, "id")
	if _, err := suite.NewMycelium(h.gateway).GetArtifact(r.Context(), userID, where, id); err != nil {
		httpjson.WriteError(w, routeError(err))
		return
	}
	share, err := h.sharing.Store.Create(r.Context(), userID, id, where)
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to create the share", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, h.shareOf(share, r))
}

func (h *Handler) getShare(w http.ResponseWriter, r *http.Request) {
	userID, err := h.shareCaller(r)
	if err == nil {
		var share shares.Share
		share, err = h.sharing.Store.Live(r.Context(), userID, chi.URLParam(r, "id"), shareSpace(r))
		if err == nil {
			httpjson.WriteJSON(w, http.StatusOK, h.shareOf(share, r))
			return
		}
	}
	httpjson.WriteError(w, shareError(err))
}

func (h *Handler) revokeShare(w http.ResponseWriter, r *http.Request) {
	userID, err := h.shareCaller(r)
	if err == nil {
		err = h.sharing.Store.Revoke(r.Context(), userID, chi.URLParam(r, "id"), shareSpace(r))
	}
	if err != nil {
		httpjson.WriteError(w, shareError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) shareCaller(r *http.Request) (int64, error) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		return 0, troncerrors.Unauthorized("not authenticated")
	}
	return identity.UserID, nil
}

func shareSpace(r *http.Request) string {
	return strings.TrimSpace(r.URL.Query().Get("mycelium_space"))
}

func shareError(err error) error {
	var known *troncerrors.Error
	switch {
	case errors.As(err, &known):
		return known
	case errors.Is(err, shares.ErrNotFound):
		return troncerrors.NotFound("no public link")
	}
	return troncerrors.Internal("failed to read the share", err)
}
