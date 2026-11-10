package suite

import (
	"context"
	"fmt"
	"net/http"
)

// ListAllArtifacts lists every artifact the user can read in Mycelium, newest
// first: their common tree when they are an admin, then each of their spaces.
// Mycelium decides what the user may read, so no other user's private
// artifact can appear here.
func (m Mycelium) ListAllArtifacts(ctx context.Context, userID int64) ([]ArtifactSummary, error) {
	raw, err := m.gw.Do(ctx, userID, Request{App: AppMycelium, Method: http.MethodGet, Path: myceliumArtifactsPath + "?all=true"})
	if err != nil {
		return nil, err
	}
	out := []ArtifactSummary{}
	if err := decodeInto(raw, &out); err != nil {
		return nil, fmt.Errorf("mycelium: reading artifacts: %w", err)
	}
	for i := range out {
		if out[i].SpaceID == "" {
			out[i].SpaceID = CommonTree
		}
	}
	return out, nil
}

// PersonalArtifacts answers where the user's personal artifacts live: the
// common tree (CommonTree) when Mycelium offers it, which it does for an admin whose
// CLI machines sync that tree, else the personal space.
func (m Mycelium) PersonalArtifacts(ctx context.Context, userID int64, personal MemorySpace) (string, error) {
	listing, err := m.spaceListing(ctx, userID)
	if err != nil {
		return "", err
	}
	if listing.CommonArtifacts {
		return CommonTree, nil
	}
	return personal.resolve(ctx)
}
