package suite

import (
	"context"
)

// ArtifactMatch is one artifact a search found. Match is "title" when the
// title holds every word, else "content", and Snippet is the body around the
// first word found.
type ArtifactMatch struct {
	ArtifactSummary
	Match   string `json:"match"`
	Snippet string `json:"snippet"`
}

// ArtifactSearchPage is one page of an artifact search. NextCursor is empty
// on the last page.
type ArtifactSearchPage struct {
	Results    []ArtifactMatch `json:"results"`
	Total      int             `json:"total"`
	NextCursor string          `json:"next_cursor,omitempty"`
}

// SearchArtifacts searches every artifact the user can read in Mycelium, by
// title and by content, one page at a time: title matches first, then
// content matches, newest first in each group. Mycelium decides what the
// user may read, as ListAllArtifacts does.
func (m Mycelium) SearchArtifacts(ctx context.Context, userID int64, query, cursor string) (ArtifactSearchPage, error) {
	var page ArtifactSearchPage
	args := map[string]any{"query": query, "cursor": cursor}
	if err := m.call(ctx, userID, "mycelium_search_artifacts", args, &page); err != nil {
		return ArtifactSearchPage{}, err
	}
	if page.Results == nil {
		page.Results = []ArtifactMatch{}
	}
	for i := range page.Results {
		if page.Results[i].SpaceID == "" {
			page.Results[i].SpaceID = CommonTree
		}
	}
	return page, nil
}
