package store

import (
	"context"
	"encoding/json"
	"time"
)

// UsageRun is one run's spend, with the conversation details a usage summary
// groups it by.
type UsageRun struct {
	StartedAt time.Time
	Channel   string
	Model     string
	Usage     json.RawMessage
	CostUSD   float64
}

type usageRow struct {
	StartedAt time.Time
	Channel   string
	Model     string
	Usage     []byte
	CostUSD   float64
}

// UsageRuns lists the user's runs that started at or after since.
func (p *Postgres) UsageRuns(ctx context.Context, userID int64, since time.Time) ([]UsageRun, error) {
	var rows []usageRow
	err := p.db.WithContext(ctx).Raw(
		`SELECT r.started_at, s.channel, s.model, r.usage, r.cost_usd::float8 AS cost_usd
		FROM runs r JOIN sessions s ON s.id = r.session_id
		WHERE s.user_id = ? AND r.started_at >= ? ORDER BY r.started_at`,
		userID, since,
	).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]UsageRun, len(rows))
	for i, row := range rows {
		out[i] = UsageRun{StartedAt: row.StartedAt, Channel: row.Channel, Model: row.Model, Usage: row.Usage, CostUSD: row.CostUSD}
	}
	return out, nil
}

// UsageRuns is the in-memory twin of Postgres.UsageRuns.
func (m *Memory) UsageRuns(_ context.Context, userID int64, since time.Time) ([]UsageRun, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []UsageRun
	for _, run := range m.runs {
		session, ok := m.sessions[run.SessionID]
		if !ok || session.UserID != userID || run.StartedAt.Before(since) {
			continue
		}
		out = append(out, UsageRun{StartedAt: run.StartedAt, Channel: session.Channel, Model: session.Model, Usage: run.Usage, CostUSD: run.CostUSD})
	}
	return out, nil
}
