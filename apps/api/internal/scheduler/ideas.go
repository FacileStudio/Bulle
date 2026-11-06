package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

// tickIdeas expires the ideas past their date, then starts the daily batch of
// every enabled space whose local hour has come and that has none yet today.
// A batch that cannot start, such as when the user already runs as many runs
// as allowed, is tried again on the next tick of the same hour.
func (s *Scheduler) tickIdeas(ctx context.Context, now time.Time) {
	if err := s.cfg.Store.ExpireStaleIdeas(ctx, now); err != nil {
		s.cfg.Logger.Error("scheduler expiring ideas failed", slog.Any("error", err))
	}
	if s.cfg.Ideas == nil {
		return
	}
	due, err := s.cfg.Store.ListEnabledIdeaSettings(ctx)
	if err != nil {
		s.cfg.Logger.Error("scheduler listing idea settings failed", slog.Any("error", err))
		return
	}
	for _, settings := range due {
		date, ok := ideasDue(settings, now)
		if !ok {
			continue
		}
		marked, err := s.cfg.Store.MarkIdeaBatchDate(ctx, settings.UserID, settings.SpaceID, date)
		if err != nil || !marked {
			continue
		}
		if _, err := s.cfg.Ideas.GenerateIdeas(ctx, settings.UserID, settings.SpaceID, store.BatchDaily); err != nil {
			s.cfg.Logger.Warn("scheduler daily ideas failed", slog.String("space_id", settings.SpaceID), slog.Any("error", err))
			if err := s.cfg.Store.ReleaseIdeaBatchDate(ctx, settings.UserID, settings.SpaceID, date); err != nil {
				s.cfg.Logger.Warn("scheduler releasing the ideas date failed", slog.String("space_id", settings.SpaceID), slog.Any("error", err))
			}
		}
	}
}

func ideasDue(settings store.IdeaSettings, now time.Time) (string, bool) {
	if !settings.Enabled {
		return "", false
	}
	loc, err := time.LoadLocation(settings.Timezone)
	if err != nil {
		loc = time.UTC
	}
	local := now.In(loc)
	date := local.Format(time.DateOnly)
	return date, local.Hour() == settings.Hour && settings.LastBatchDate != date
}
