package runner

import (
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

// watchMemory marks the chat read by memory once its pass finished. A pass
// that failed or stopped leaves it unread, so the next pass learns it again.
func (r *Runner) watchMemory(run *memoryRun, runID string) {
	defer r.releaseMemory(run.userID)
	defer r.forgetMemory(run)
	if !r.awaitEnd(run.userID, runID) {
		return
	}
	if ended, err := r.cfg.Store.GetRun(r.base, run.userID, runID); err != nil || ended.Status != store.RunFinished {
		r.cfg.Logger.Warn("memory pass did not finish; the chat stays unread by memory", "session", run.source, "status", ended.Status, "error", err)
		return
	}
	if err := r.cfg.Store.MarkMemorySeen(r.base, run.source, run.upto); err != nil {
		r.cfg.Logger.Warn("marking the chat seen by memory", "session", run.source, "error", err)
	}
}

// awaitEnd blocks until the run is over and answers true, or answers false
// when the runner shuts down first.
func (r *Runner) awaitEnd(userID int64, runID string) bool {
	notify, unsubscribe := r.Subscribe(runID)
	defer unsubscribe()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for r.runOngoing(userID, runID) {
		select {
		case <-r.base.Done():
			return false
		case <-notify:
		case <-ticker.C:
		}
	}
	return true
}

func (r *Runner) runOngoing(userID int64, runID string) bool {
	if r.IsLive(runID) {
		return true
	}
	found, err := r.cfg.Store.GetRun(r.base, userID, runID)
	return err == nil && found.Status == store.RunRunning
}
