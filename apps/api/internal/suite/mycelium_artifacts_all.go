package suite

import (
	"context"
)

// ListAllArtifacts lists every artifact the user can read in Mycelium, newest
// first: their common tree when they are an admin, then each of their spaces.
// Mycelium decides what the user may read, so no other user's private
// artifact can appear here.
func (m Mycelium) ListAllArtifacts(ctx context.Context, userID int64) ([]ArtifactSummary, error) {
	out, err := m.listArtifacts(ctx, userID, map[string]any{})
	if err != nil {
		return nil, err
	}
	for i := range out {
		if out[i].SpaceID == "" {
			out[i].SpaceID = CommonTree
		}
	}
	return out, nil
}

// PersonalArtifacts answers where the user's personal artifacts live: the
// common tree (CommonTree) when Mycelium offers it, which it does for an
// admin, else the personal space.
func (m Mycelium) PersonalArtifacts(ctx context.Context, userID int64, personal MemorySpace) (string, error) {
	trees, err := m.trees(ctx, userID)
	if err != nil {
		return "", err
	}
	for _, tree := range trees {
		if tree.SpaceID == "" {
			return CommonTree, nil
		}
	}
	return personal.resolve(ctx)
}
