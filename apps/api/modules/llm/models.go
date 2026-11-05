package llm

import (
	"context"
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/catalog"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

// CatalogFunc is the catalogue bound to the instance's backend.
type CatalogFunc func(ctx context.Context, userID int64) []catalog.Model

// CatalogResponse is what GET /models returns: the same list /api/llm puts in
// "models", with the vendor's display names for a picker to search.
type CatalogResponse struct {
	Models []catalog.Model `json:"models"`
}

func catalogHandler(lister CatalogFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, ok := authcontext.IdentityFromContext(r.Context())
		if !ok {
			httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
			return
		}
		models := []catalog.Model{}
		if lister != nil {
			models = lister(r.Context(), identity.UserID)
		}
		httpjson.WriteJSON(w, http.StatusOK, CatalogResponse{Models: models})
	}
}
