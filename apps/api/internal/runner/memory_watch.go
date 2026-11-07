package runner

import (
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

func (r *Runner) watchMemory(run *memoryRun, runID string) {
	defer r.releaseMemory(run.userID)
	defer r.forgetMemory(run)
	if !r.awaitEnd(run.userID, runID) {
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
