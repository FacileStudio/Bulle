package memory

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/go-chi/chi/v5"
)

// resolve answers the caller and the Mycelium memory space of the Bulle space
// named by space_id, or of the default space when it is omitted. The Bulle
// space is looked up among the caller's own, so another user's id is refused
// before Mycelium is reached.
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

// target is one page a route acts on: the caller, its memory space, and the
// page path as given and as escaped for Mycelium.
type target struct {
	userID  int64
	space   string
	path    string
	escaped string
}

// target checks the page the route names before Mycelium is reached: relative
// to the memory directory, no hidden or ".." segment, and markdown.
func (h *Handler) target(r *http.Request) (target, error) {
	raw, err := url.PathUnescape(chi.URLParam(r, "*"))
	if err != nil {
		return target{}, troncerrors.Invalid("invalid page path")
	}
	escaped, err := suite.MemoryPath(raw, true)
	if err != nil {
		return target{}, troncerrors.Invalid(err.Error())
	}
	userID, space, err := h.resolve(r)
	return target{userID: userID, space: space, path: strings.TrimSpace(raw), escaped: escaped}, err
}
