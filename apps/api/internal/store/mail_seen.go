package store

import (
	"context"
	"time"
)

// MailSeenFor is how long Bulle remembers that the assistant saw an email.
const MailSeenFor = 60 * 24 * time.Hour

// MailSeen remembers which Courrier emails the assistant was already shown,
// apart from the read flag, which belongs to the user.
type MailSeen interface {
	SeeMail(ctx context.Context, userID int64, emailIDs []int64, now time.Time) (map[int64]time.Time, error)
}

// SeeMail returns when the assistant first saw each of emailIDs it had seen
// before now, records the rest as seen now, and forgets sightings older than
// MailSeenFor.
func (p *Postgres) SeeMail(ctx context.Context, userID int64, emailIDs []int64, now time.Time) (map[int64]time.Time, error) {
	seen := map[int64]time.Time{}
	if len(emailIDs) == 0 {
		return seen, nil
	}
	db := p.db.WithContext(ctx)
	if err := db.Exec(`DELETE FROM mail_seen WHERE user_id = ? AND seen_at < ?`, userID, now.Add(-MailSeenFor)).Error; err != nil {
		return nil, err
	}
	var rows []struct {
		EmailID int64
		SeenAt  time.Time
	}
	if err := db.Raw(`SELECT email_id, seen_at FROM mail_seen WHERE user_id = ? AND email_id IN ?`, userID, emailIDs).Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		seen[row.EmailID] = row.SeenAt
	}
	for _, id := range emailIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		if err := db.Exec(`INSERT INTO mail_seen (user_id, email_id, seen_at) VALUES (?, ?, ?) ON CONFLICT DO NOTHING`, userID, id, now).Error; err != nil {
			return nil, err
		}
	}
	return seen, nil
}

// SeeMail is the in-memory twin of Postgres.SeeMail.
func (m *Memory) SeeMail(_ context.Context, userID int64, emailIDs []int64, now time.Time) (map[int64]time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.mailSeen == nil {
		m.mailSeen = map[mailKey]time.Time{}
	}
	seen := map[int64]time.Time{}
	for _, id := range emailIDs {
		key := mailKey{userID: userID, emailID: id}
		at, ok := m.mailSeen[key]
		switch {
		case ok && now.Sub(at) < MailSeenFor:
			seen[id] = at
		default:
			m.mailSeen[key] = now
		}
	}
	return seen, nil
}

type mailKey struct {
	userID  int64
	emailID int64
}
