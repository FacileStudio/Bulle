package artifacts

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/shares"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

// PublicOrigin answers the scheme and host of the instance's public URL, or
// an empty string when raw is not an absolute URL.
func PublicOrigin(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host
}

func (h *Handler) publicShare(w http.ResponseWriter, r *http.Request) {
	header := w.Header()
	header.Set("Cache-Control", "no-store")
	header.Set("X-Robots-Tag", "noindex, nofollow")
	header.Set("Referrer-Policy", "no-referrer")
	share, err := h.sharing.Store.Lookup(r.Context(), chi.URLParam(r, "token"))
	if errors.Is(err, shares.ErrNotFound) {
		httpjson.WriteError(w, troncerrors.NotFound("share not found"))
		return
	}
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to read the share", err))
		return
	}
	found, err := suite.NewMycelium(h.gateway).GetArtifact(r.Context(), share.UserID, share.Space, share.ArtifactID)
	if err != nil {
		httpjson.WriteError(w, publicError(err))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, PublicShareResponse{Title: found.Title, Content: found.Content, Updated: found.Updated})
}

func publicError(err error) error {
	var refused *suite.RefusedError
	if errors.As(err, &refused) && refused.Status == http.StatusNotFound {
		return troncerrors.NotFound("share not found")
	}
	return troncerrors.Unavailable("this share is not available right now")
}
