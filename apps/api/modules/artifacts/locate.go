package artifacts

import (
	"net/http"
	"slices"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) caller(r *http.Request) (int64, error) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		return 0, troncerrors.Unauthorized("not authenticated")
	}
	if h.gateway == nil || !h.gateway.Config().MyceliumEnabled() {
		return identity.UserID, errNotConnected
	}
	return identity.UserID, nil
}

// locate answers the Mycelium tree that holds the artifact named in the path.
// The mycelium_space query names it outright, as the list hands it out.
// Without it, the artifact is looked up among everything the user can read,
// so a chat card opens whichever space or machine wrote it. When the id is in
// more than one tree, the one the space_id space writes to wins. Mycelium
// checks every call against the user, so naming a tree grants nothing.
func (h *Handler) locate(r *http.Request) (int64, string, error) {
	userID, err := h.caller(r)
	if err != nil {
		return userID, "", err
	}
	if where := strings.TrimSpace(r.URL.Query().Get("mycelium_space")); where != "" {
		return userID, where, nil
	}
	mycelium := suite.NewMycelium(h.gateway)
	items, err := mycelium.ListAllArtifacts(r.Context(), userID)
	if err != nil {
		return userID, "", err
	}
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	trees := []string{}
	for _, item := range items {
		if item.ID == id {
			trees = append(trees, item.SpaceID)
		}
	}
	switch len(trees) {
	case 0:
		return userID, "", troncerrors.NotFound("artifact not found")
	case 1:
		return userID, trees[0], nil
	}
	space, err := h.space(r.Context(), userID, strings.TrimSpace(r.URL.Query().Get("space_id")))
	if err != nil {
		return userID, "", err
	}
	preferred, err := mycelium.ArtifactSpaceOf(r.Context(), userID, space, h.store)
	if err == nil && slices.Contains(trees, preferred) {
		return userID, preferred, nil
	}
	return userID, trees[0], nil
}
