package runner

import (
	"context"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

// runSlot is the key a run reserves under: the owner's id, or its negation for
// a guest session. Colleagues writing to the owner's bot share a pool of run
// slots of their own, so they never hold the owner's.
func runSlot(userID int64, channel string) int64 {
	if channel == store.ChannelGuest {
		return -userID
	}
	return userID
}

func (r *Runner) slotFor(ctx context.Context, userID int64, sessionID string) (int64, error) {
	session, err := r.cfg.Store.GetSession(ctx, userID, sessionID)
	if err != nil {
		return 0, err
	}
	return runSlot(userID, session.Channel), nil
}
