package matrixbot

import (
	"log/slog"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/prisme"
)

// retryTaken is how long a user whose subject belongs to another Prisme
// account is left alone before the loop tries again.
const retryTaken = 24 * time.Hour

// onHold reports whether a refusal still keeps the user out of the loop: for
// good (a zero expiry) or until the expiry passes.
func (m *Manager) onHold(userID int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	until, held := m.held[userID]
	return held && (until.IsZero() || m.now().Before(until))
}

// hold leaves the user out of the provisioning loop. acting_subject_taken
// heals by itself once the person signs into Prisme with their new address,
// so it is retried after a day and logged at WARN each time it fails again.
// Every other final refusal stands until the next start and is logged once,
// at ERROR, with Prisme's code for the operator.
func (m *Manager) hold(userID int64, r *prisme.Refusal) {
	m.mu.Lock()
	defer m.mu.Unlock()
	attrs := []any{slog.Int64("user_id", userID), slog.String("code", r.Code), slog.Int("status", r.Status), slog.String("message", r.Message)}
	if r.Code == prisme.CodeSubjectTaken {
		m.held[userID] = m.now().Add(retryTaken)
		m.cfg.Logger.Warn("matrix: prisme holds the subject for another account, retried in a day", attrs...)
		return
	}
	if until, held := m.held[userID]; held && until.IsZero() {
		return
	}
	m.held[userID] = time.Time{}
	m.cfg.Logger.Error("matrix: prisme refused the user's identity, not retried until restart", attrs...)
}
