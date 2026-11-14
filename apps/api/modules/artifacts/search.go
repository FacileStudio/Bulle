package artifacts

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

const maxQuery = 200

func (h *Handler) searchArtifacts(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" || utf8.RuneCountInString(query) > maxQuery {
		httpjson.WriteError(w, troncerrors.Invalid("q must hold 1 to 200 characters"))
		return
	}
	userID, err := h.caller(r)
	if err == nil {
		var page suite.ArtifactSearchPage
		page, err = suite.NewMycelium(h.gateway).SearchArtifacts(r.Context(), userID, query, r.URL.Query().Get("cursor"))
		if err == nil {
			httpjson.WriteJSON(w, http.StatusOK, SearchResponse{Artifacts: toHits(page.Results), NextCursor: page.NextCursor, Available: true})
			return
		}
	}
	if disconnected(err) {
		httpjson.WriteJSON(w, http.StatusOK, SearchResponse{Artifacts: []SearchHit{}, Available: false})
		return
	}
	httpjson.WriteError(w, routeError(err))
}

func toHits(items []suite.ArtifactMatch) []SearchHit {
	out := make([]SearchHit, 0, len(items))
	for _, item := range items {
		out = append(out, SearchHit{Artifact: toArtifact(item.ArtifactSummary), Match: item.Match, Snippet: item.Snippet})
	}
	return out
}
