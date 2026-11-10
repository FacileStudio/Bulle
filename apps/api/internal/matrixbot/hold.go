package matrixbot

import (
	"log/slog"

	"github.com/FacileStudio/Bulle/apps/api/internal/prisme"
)

// onHold reports whether Prisme gave the user a refusal no retry can change.
func (m *Manager) onHold(userID int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, held := m.held[userID]
	return held
}

// hold leaves the user out of the provisioning loop until the next start and
// logs the refusal once, with Prisme's error code, for the operator.
func (m *Manager) hold(userID int64, r *prisme.Refusal) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, held := m.held[userID]; held {
		return
	}
	m.held[userID] = r.Code
	m.cfg.Logger.Error("matrix: prisme refused the user's identity, not retried until restart",
		slog.Int64("user_id", userID), slog.String("code", r.Code), slog.Int("status", r.Status), slog.String("message", r.Message))
}
