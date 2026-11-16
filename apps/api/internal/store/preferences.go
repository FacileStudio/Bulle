package store

import (
	"context"
	"encoding/json"
	"time"
)

// GetPreferences answers what the user chose for themselves, or ErrNotFound
// when they never saved a choice.
func (m *Memory) GetPreferences(_ context.Context, userID int64) (Preferences, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	prefs, ok := m.prefs[userID]
	if !ok {
		return Preferences{}, ErrNotFound
	}
	return prefs, nil
}

func (m *Memory) SavePreferences(_ context.Context, prefs Preferences) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	prefs.UpdatedAt = time.Now().UTC()
	prefs.ComputerNetwork = m.prefs[prefs.UserID].ComputerNetwork
	prefs.ComputerNetAllow = m.prefs[prefs.UserID].ComputerNetAllow
	m.prefs[prefs.UserID] = prefs
	return nil
}

type preferencesRow struct {
	UserID          int64
	Model           string
	ComputerNetwork string
	NetAllow        string
	UpdatedAt       time.Time
}

func (p *Postgres) GetPreferences(ctx context.Context, userID int64) (Preferences, error) {
	var rows []preferencesRow
	err := p.db.WithContext(ctx).Raw(
		"SELECT user_id, model, computer_network, computer_net_allow::text AS net_allow, updated_at FROM user_preferences WHERE user_id = ?", userID,
	).Scan(&rows).Error
	if err != nil {
		return Preferences{}, err
	}
	if len(rows) == 0 {
		return Preferences{}, ErrNotFound
	}
	row := rows[0]
	var allow []string
	if row.NetAllow != "" {
		if err := json.Unmarshal([]byte(row.NetAllow), &allow); err != nil {
			return Preferences{}, err
		}
	}
	return Preferences{UserID: row.UserID, Model: row.Model, ComputerNetwork: row.ComputerNetwork, ComputerNetAllow: allow, UpdatedAt: row.UpdatedAt}, nil
}

func (p *Postgres) SavePreferences(ctx context.Context, prefs Preferences) error {
	return p.db.WithContext(ctx).Exec(`
		INSERT INTO user_preferences (user_id, model, created_at, updated_at)
		VALUES (?, ?, now(), now())
		ON CONFLICT (user_id) DO UPDATE SET
			model = EXCLUDED.model,
			updated_at = now()
	`, prefs.UserID, prefs.Model).Error
}
