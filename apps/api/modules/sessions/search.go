package sessions

import (
	"errors"
	"net/http"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

// SearchResponse is what GET /sessions/search answers: the caller's web and
// matrix chats of one space that match, newest first, at most 20.
type SearchResponse struct {
	Results []store.SessionHit `json:"results"`
}

func (h *Handler) searchSessions(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	space, err := h.searchSpace(r, identity.UserID)
	if errors.Is(err, store.ErrNotFound) {
		httpjson.WriteError(w, troncerrors.NotFound("space not found"))
		return
	}
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to load the space", err))
		return
	}
	hits, err := runner.SearchChats(r.Context(), h.store, identity.UserID, space.ID, r.URL.Query().Get("q"))
	if errors.Is(err, runner.ErrShortQuery) {
		httpjson.WriteError(w, troncerrors.Invalid(err.Error()))
		return
	}
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to search sessions", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, SearchResponse{Results: hits})
}

func (h *Handler) searchSpace(r *http.Request, userID int64) (store.Space, error) {
	spaceID := strings.TrimSpace(r.URL.Query().Get("space_id"))
	if spaceID == "" {
		return h.store.EnsureDefaultSpace(r.Context(), userID)
	}
	return h.store.GetSpace(r.Context(), userID, spaceID)
}
