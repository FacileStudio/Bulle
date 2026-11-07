package store

import (
	"context"
	stderrors "errors"
	"time"
)

// NetworkOpen is the computer network mode a person who never chose reads as.
const NetworkOpen = "open"

// ComputerNetworks keeps the egress mode, open or offline, each person chose
// for their computer, apart from the rest of their preferences so saving one
// never overwrites the other.
type ComputerNetworks interface {
	ComputerNetwork(ctx context.Context, userID int64) (string, error)
	SaveComputerNetwork(ctx context.Context, userID int64, network string) error
}

// ComputerNetwork returns the person's computer network mode, NetworkOpen when
// they never chose one.
func (m *Memory) ComputerNetwork(ctx context.Context, userID int64) (string, error) {
	return networkOf(m.GetPreferences(ctx, userID))
}

// SaveComputerNetwork records the person's computer network mode.
func (m *Memory) SaveComputerNetwork(_ context.Context, userID int64, network string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	prefs := m.prefs[userID]
	prefs.UserID, prefs.ComputerNetwork, prefs.UpdatedAt = userID, network, time.Now().UTC()
	m.prefs[userID] = prefs
	return nil
}

// ComputerNetwork returns the person's computer network mode, NetworkOpen when
// they never chose one.
func (p *Postgres) ComputerNetwork(ctx context.Context, userID int64) (string, error) {
	return networkOf(p.GetPreferences(ctx, userID))
}

// SaveComputerNetwork records the person's computer network mode and leaves
// their model choice alone.
func (p *Postgres) SaveComputerNetwork(ctx context.Context, userID int64, network string) error {
	return p.db.WithContext(ctx).Exec(`
		INSERT INTO user_preferences (user_id, computer_network, created_at, updated_at)
		VALUES (?, ?, now(), now())
		ON CONFLICT (user_id) DO UPDATE SET
			computer_network = EXCLUDED.computer_network,
			updated_at = now()
	`, userID, network).Error
}

func networkOf(prefs Preferences, err error) (string, error) {
	if stderrors.Is(err, ErrNotFound) {
		return NetworkOpen, nil
	}
	if err != nil {
		return "", err
	}
	if prefs.ComputerNetwork == "" {
		return NetworkOpen, nil
	}
	return prefs.ComputerNetwork, nil
}
