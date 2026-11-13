package sessions

import (
	"github.com/FacileStudio/Bulle/apps/api/internal/live"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

// summaryOf is the list row for s. Its status comes from the runner's runs in
// flight: running, waiting for an approval, or idle when it has none.
func summaryOf(s store.Session, statuses map[string]string) SessionSummary {
	status, ok := statuses[s.ID]
	if !ok {
		status = live.StatusIdle
	}
	return SessionSummary{
		ID:        s.ID,
		SpaceID:   s.SpaceID,
		Title:     s.Title,
		Channel:   s.Channel,
		UpdatedAt: s.UpdatedAt.UTC(),
		Status:    status,
		Unread:    s.Unread(),
	}
}
