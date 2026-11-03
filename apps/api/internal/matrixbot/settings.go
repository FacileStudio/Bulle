package matrixbot

import "context"

// Status is what the web app shows: the agent's Matrix account, the rooms it
// sits in with the space each one works in, and whether it is running. There
// is nothing to configure; Prisme provisions it all.
type Status struct {
	Available bool   `json:"available"`
	User      string `json:"user_id"`
	Rooms     []Room `json:"rooms"`
	Status    string `json:"status"`
}

// Get returns the user's agent status.
func (m *Manager) Get(ctx context.Context, userID int64) (Status, error) {
	bot, _, err := m.repo.get(ctx, userID)
	if err != nil {
		return Status{}, err
	}
	rooms, err := m.repo.rooms(ctx, userID)
	if err != nil {
		return Status{}, err
	}
	return Status{Available: m.seal != nil && m.cfg.Prisme != nil, User: bot.User, Rooms: rooms, Status: m.statusOf(userID)}, nil
}
