package computers

import (
	"context"
	"log/slog"
	"time"
)

func (s *Service) Sweep(ctx context.Context) int {
	if s.config.IdleAfter <= 0 {
		return 0
	}
	idle, err := s.store.Idle(ctx, time.Now().Add(-s.config.IdleAfter))
	if err != nil {
		s.logger.Error("idle sweep failed to list computers", slog.Any("error", err))
		return 0
	}
	stopped := 0
	for _, candidate := range idle {
		if s.sweepOne(ctx, candidate.UserID) {
			stopped++
		}
	}
	return stopped
}

func (s *Service) sweepOne(ctx context.Context, userID int64) bool {
	defer s.locks.lock(userID)()
	current, err := s.store.Find(ctx, userID)
	if err != nil || current == nil || current.Status != StatusRunning {
		return false
	}
	if time.Since(current.LastActiveAt) < s.config.IdleAfter {
		return false
	}
	if _, err := s.stopLocked(ctx, userID, false); err != nil {
		s.logger.Error("idle sweep failed to stop a computer", slog.Int64("user_id", userID), slog.Any("error", err))
		return false
	}
	s.logger.Info("stopped an idle computer", slog.Int64("user_id", userID))
	return true
}

func (s *Service) RunSweeper(ctx context.Context) {
	if s.config.IdleAfter <= 0 {
		return
	}
	interval := min(s.config.IdleAfter/2, time.Minute)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.Sweep(ctx)
		}
	}
}
