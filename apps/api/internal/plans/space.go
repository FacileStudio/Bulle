package plans

import (
	"context"
	"errors"

	"github.com/FacileStudio/porte/forfait"
	troncerrors "github.com/FacileStudio/tronc/errors"
)

// CheckAIIn is CheckAI for a request made for a Registre shared space: it
// also answers nil when spaceID's plan has ai.access and the user is a member
// of that space, as Bulle's mirror of Registre's spaces records it. A space
// the user is not a member of counts for nothing, so naming an id never
// grants anything by itself. An empty spaceID is CheckAI.
func (g *Gate) CheckAIIn(ctx context.Context, userID int64, spaceID string) error {
	refusal := g.CheckAI(ctx, userID)
	if refusal == nil || spaceID == "" || !errors.Is(refusal, ErrAIAccess) {
		return refusal
	}
	if !g.store.Enabled(forfait.Scope{SpaceID: spaceID}, AIAccess) {
		return refusal
	}
	var members int64
	if err := g.orm.WithContext(ctx).Raw(`SELECT count(*) FROM spaces WHERE user_id = ? AND registre_space_id = ?`,
		userID, spaceID).Scan(&members).Error; err != nil {
		return troncerrors.Internal("failed to read the user's spaces", err)
	}
	if members == 0 {
		return refusal
	}
	return nil
}
