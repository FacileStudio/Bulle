package runner

import "github.com/FacileStudio/Bulle/apps/api/internal/live"

// Statuses maps each of the user's sessions with a run in flight to running,
// or to waiting while a tool call waits for approval. Idle sessions are absent.
func (r *Runner) Statuses(userID int64) map[string]string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := map[string]string{}
	for _, lr := range r.live {
		if lr.userID == userID {
			out[lr.session.ID] = lr.status()
		}
	}
	return out
}

func (r *Runner) sessionStatus(userID int64, sessionID string) string {
	if status, ok := r.Statuses(userID)[sessionID]; ok {
		return status
	}
	return live.StatusIdle
}

func (lr *liveRun) status() string {
	lr.gate.mu.Lock()
	defer lr.gate.mu.Unlock()
	if len(lr.gate.waiting) > 0 {
		return live.StatusWaiting
	}
	return live.StatusRunning
}
