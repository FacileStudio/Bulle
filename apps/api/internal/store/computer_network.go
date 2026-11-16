package store

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"slices"
	"time"
)

// ComputerNetworks keeps the egress mode each person chose for their computer
// (open, offline or allowlist) and, for allowlist, their extra hosts, apart
// from the rest of their preferences so saving one never overwrites the other.
// An empty mode means the person never chose.
type ComputerNetworks interface {
	ComputerNetwork(ctx context.Context, userID int64) (mode string, allow []string, err error)
	SaveComputerNetwork(ctx context.Context, userID int64, mode string, allow []string) error
}

// ComputerNetwork returns the person's computer network mode and extra hosts,
// an empty mode when they never chose one.
func (m *Memory) ComputerNetwork(ctx context.Context, userID int64) (string, []string, error) {
	return networkOf(m.GetPreferences(ctx, userID))
}

// SaveComputerNetwork records the person's computer network mode and extra
// hosts.
func (m *Memory) SaveComputerNetwork(_ context.Context, userID int64, mode string, allow []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	prefs := m.prefs[userID]
	prefs.UserID, prefs.ComputerNetwork, prefs.ComputerNetAllow, prefs.UpdatedAt = userID, mode, slices.Clone(allow), time.Now().UTC()
	m.prefs[userID] = prefs
	return nil
}

// ComputerNetwork returns the person's computer network mode and extra hosts,
// an empty mode when they never chose one.
func (p *Postgres) ComputerNetwork(ctx context.Context, userID int64) (string, []string, error) {
	return networkOf(p.GetPreferences(ctx, userID))
}

// SaveComputerNetwork records the person's computer network mode and extra
// hosts and leaves their model choice alone.
func (p *Postgres) SaveComputerNetwork(ctx context.Context, userID int64, mode string, allow []string) error {
	encoded, err := json.Marshal(nonNil(allow))
	if err != nil {
		return err
	}
	return p.db.WithContext(ctx).Exec(`
		INSERT INTO user_preferences (user_id, computer_network, computer_net_allow, created_at, updated_at)
		VALUES (?, ?, ?::jsonb, now(), now())
		ON CONFLICT (user_id) DO UPDATE SET
			computer_network = EXCLUDED.computer_network,
			computer_net_allow = EXCLUDED.computer_net_allow,
			updated_at = now()
	`, userID, mode, string(encoded)).Error
}

func nonNil(hosts []string) []string {
	if hosts == nil {
		return []string{}
	}
	return hosts
}

func networkOf(prefs Preferences, err error) (string, []string, error) {
	if stderrors.Is(err, ErrNotFound) {
		return "", nil, nil
	}
	if err != nil {
		return "", nil, err
	}
	return prefs.ComputerNetwork, prefs.ComputerNetAllow, nil
}
