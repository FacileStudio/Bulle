package runner

import (
	"context"
	"slices"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

// CreateWebSession opens a web chat for the user in spaceID, or in the
// default space when it is empty, on the instance backend and the user's
// saved model when the instance still offers it.
func (r *Runner) CreateWebSession(ctx context.Context, userID int64, spaceID, title string) (store.Session, error) {
	backend, model := r.Defaults()
	if prefs, err := r.cfg.Store.GetPreferences(ctx, userID); err == nil {
		if prefs.Model != "" && slices.Contains(r.Models(), prefs.Model) {
			model = prefs.Model
		}
	}
	return r.cfg.Store.CreateSession(ctx, store.Session{
		ID: store.NewID(), UserID: userID, SpaceID: spaceID, Channel: store.ChannelWeb,
		Title: title, Backend: backend, Model: model,
	})
}
