package scheduler

import (
	"context"
	"log/slog"
	"time"
)

const (
	// MemoryIdle is how long a chat stays quiet before its memory pass.
	MemoryIdle = 10 * time.Minute
	// MemoryPassesPerTick caps the memory passes one tick starts.
	MemoryPassesPerTick = 5
)

// MemoryPasser learns from idle chats; *runner.Runner satisfies it.
type MemoryPasser interface {
	MemoryAvailable(ctx context.Context, userID int64) bool
	MemoryPass(ctx context.Context, userID int64, sessionID string) error
}

// tickMemory starts a memory pass for chats idle ten minutes with messages no
// pass has read, at most five per tick. A user without Mycelium is skipped
// and asked about once per tick; a pass that cannot start now, such as while
// the user already has one running, is tried again on a later tick.
func (s *Scheduler) tickMemory(ctx context.Context, now time.Time) {
	if s.cfg.Memory == nil {
		return
	}
	due, err := s.cfg.Store.ListSessionsForMemoryPass(ctx, now.Add(-MemoryIdle))
	if err != nil {
		s.cfg.Logger.Error("scheduler listing chats for memory failed", slog.Any("error", err))
		return
	}
	available := map[int64]bool{}
	started := 0
	for _, session := range due {
		if started >= MemoryPassesPerTick {
			return
		}
		if !s.memoryAvailable(ctx, available, session.UserID) {
			continue
		}
		if err := s.cfg.Memory.MemoryPass(ctx, session.UserID, session.ID); err != nil {
			s.cfg.Logger.Debug("scheduler memory pass not started", slog.String("session_id", session.ID), slog.Any("error", err))
			continue
		}
		started++
	}
}

func (s *Scheduler) memoryAvailable(ctx context.Context, asked map[int64]bool, userID int64) bool {
	ok, known := asked[userID]
	if !known {
		ok = s.cfg.Memory.MemoryAvailable(ctx, userID)
		asked[userID] = ok
	}
	return ok
}
