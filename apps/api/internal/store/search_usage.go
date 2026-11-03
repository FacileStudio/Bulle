package store

import (
	"context"
	"time"
)

// SearchQuota counts the web searches each user runs on the instance's search
// keys, one row per user and UTC day.
type SearchQuota interface {
	TakeSearch(ctx context.Context, userID int64, day time.Time, limit int) (bool, error)
}

// TakeSearch counts one search for the user on day and reports true, or
// reports false without counting when the day already holds limit searches.
func (p *Postgres) TakeSearch(ctx context.Context, userID int64, day time.Time, limit int) (bool, error) {
	if limit <= 0 {
		return false, nil
	}
	var counts []int
	err := p.db.WithContext(ctx).Raw(
		`INSERT INTO search_usage (user_id, day, count) VALUES (?, ?, 1)
		ON CONFLICT (user_id, day) DO UPDATE SET count = search_usage.count + 1
		WHERE search_usage.count < ?
		RETURNING count`,
		userID, day.UTC().Format(time.DateOnly), limit,
	).Scan(&counts).Error
	if err != nil {
		return false, err
	}
	return len(counts) > 0, nil
}

// TakeSearch is the in-memory twin of Postgres.TakeSearch.
func (m *Memory) TakeSearch(_ context.Context, userID int64, day time.Time, limit int) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.searches == nil {
		m.searches = map[searchDay]int{}
	}
	key := searchDay{userID: userID, day: day.UTC().Format(time.DateOnly)}
	if m.searches[key] >= limit {
		return false, nil
	}
	m.searches[key]++
	return true, nil
}

type searchDay struct {
	userID int64
	day    string
}
