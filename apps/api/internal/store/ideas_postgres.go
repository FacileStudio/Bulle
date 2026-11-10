package store

import (
	"context"
	"encoding/json"
	"time"
)

const ideaColumns = `id, user_id, space_id, batch_id, app, kind, title, why, prompt, draft, sources, fingerprint,
	status, feedback, reason, session_id, created_at, updated_at, expires_at`

type ideaRow struct {
	ID          string
	UserID      int64
	SpaceID     string
	BatchID     string
	App         string
	Kind        string
	Title       string
	Why         string
	Prompt      string
	Draft       string
	Sources     []byte
	Fingerprint string
	Status      string
	Feedback    string
	Reason      string
	SessionID   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ExpiresAt   time.Time
}

func (r ideaRow) idea() Idea {
	sources := []IdeaSource{}
	_ = json.Unmarshal(r.Sources, &sources)
	return Idea{
		ID: r.ID, UserID: r.UserID, SpaceID: r.SpaceID, BatchID: r.BatchID, App: r.App, Kind: r.Kind,
		Title: r.Title, Why: r.Why, Prompt: r.Prompt, Draft: r.Draft, Sources: sources,
		Fingerprint: r.Fingerprint, Status: r.Status, Feedback: r.Feedback, Reason: r.Reason,
		SessionID: r.SessionID, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt, ExpiresAt: r.ExpiresAt,
	}
}

func (p *Postgres) ideas(ctx context.Context, query string, args ...any) ([]Idea, error) {
	var rows []ideaRow
	if err := p.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]Idea, len(rows))
	for i, row := range rows {
		out[i] = row.idea()
	}
	return out, nil
}

func (p *Postgres) oneIdea(ctx context.Context, query string, args ...any) (Idea, error) {
	found, err := p.ideas(ctx, query, args...)
	if err != nil {
		return Idea{}, err
	}
	if len(found) == 0 {
		return Idea{}, ErrNotFound
	}
	return found[0], nil
}

func (p *Postgres) ListActiveIdeas(ctx context.Context, userID int64, spaceID, app string, now time.Time) ([]Idea, error) {
	return p.ideas(ctx, "SELECT "+ideaColumns+` FROM ideas WHERE user_id = ? AND space_id = ? AND (? = '' OR app = ?)
		AND (status = 'saved' OR (status = 'new' AND expires_at > ?)) ORDER BY created_at DESC, id`,
		userID, spaceID, app, app, now)
}

func (p *Postgres) GetIdea(ctx context.Context, userID int64, id string) (Idea, error) {
	return p.oneIdea(ctx, "SELECT "+ideaColumns+" FROM ideas WHERE id = ? AND user_id = ?", id, userID)
}

func (p *Postgres) CreateIdea(ctx context.Context, idea Idea) (Idea, error) {
	if idea.ID == "" {
		idea.ID = NewID()
	}
	if idea.Status == "" {
		idea.Status = IdeaNew
	}
	if idea.App == "" {
		idea.App = IdeaAppOf(idea.Fingerprint)
	}
	sources, err := json.Marshal(cloneIdea(idea).Sources)
	if err != nil {
		return Idea{}, err
	}
	return p.oneIdea(ctx, `INSERT INTO ideas (id, user_id, space_id, batch_id, app, kind, title, why, prompt, draft,
		sources, fingerprint, status, expires_at)
		SELECT ?, ?, id, ?, ?, ?, ?, ?, ?, ?, ?::jsonb, ?, ?, ? FROM spaces WHERE id = ? AND user_id = ?
		RETURNING `+ideaColumns,
		idea.ID, idea.UserID, idea.BatchID, idea.App, idea.Kind, idea.Title, idea.Why, idea.Prompt, idea.Draft,
		string(sources), idea.Fingerprint, idea.Status, idea.ExpiresAt, idea.SpaceID, idea.UserID)
}

func (p *Postgres) UpdateIdea(ctx context.Context, idea Idea) (Idea, error) {
	return p.oneIdea(ctx, `UPDATE ideas SET status = ?, feedback = ?, reason = ?, session_id = ?, updated_at = now()
		WHERE id = ? AND user_id = ? RETURNING `+ideaColumns,
		idea.Status, idea.Feedback, idea.Reason, idea.SessionID, idea.ID, idea.UserID)
}

func (p *Postgres) ExpireIdeas(ctx context.Context, userID int64, spaceID, app, keepBatchID string) error {
	scope := `AND app = ?`
	args := []any{userID, spaceID, keepBatchID, app}
	if app == "" {
		scope = `AND batch_id NOT IN (SELECT id FROM idea_batches WHERE user_id = ? AND space_id = ? AND app <> '')`
		args = []any{userID, spaceID, keepBatchID, userID, spaceID}
	}
	return p.db.WithContext(ctx).Exec(
		`UPDATE ideas SET status = 'expired', updated_at = now()
		WHERE user_id = ? AND space_id = ? AND status = 'new' AND batch_id <> ? `+scope,
		args...,
	).Error
}

func (p *Postgres) ExpireStaleIdeas(ctx context.Context, now time.Time) error {
	return p.db.WithContext(ctx).Exec(
		"UPDATE ideas SET status = 'expired', updated_at = now() WHERE status = 'new' AND expires_at <= ?", now,
	).Error
}

func (p *Postgres) IdeaFeedback(ctx context.Context, userID int64, spaceID, app string, limit int) ([]Idea, error) {
	return p.ideas(ctx, "SELECT "+ideaColumns+` FROM ideas WHERE user_id = ? AND space_id = ? AND (? = '' OR app = ?)
		AND (feedback <> '' OR status = 'dismissed') ORDER BY updated_at DESC, id LIMIT ?`,
		userID, spaceID, app, app, limit)
}

func (p *Postgres) SuppressedFingerprints(ctx context.Context, userID int64, spaceID string, now time.Time) ([]string, error) {
	var out []string
	err := p.db.WithContext(ctx).Raw(
		`SELECT DISTINCT fingerprint FROM ideas WHERE user_id = ? AND space_id = ? AND (
			status = 'saved' OR (status = 'new' AND expires_at > ?)
			OR (status IN ('dismissed', 'done') AND updated_at >= ?)
		) ORDER BY fingerprint`,
		userID, spaceID, now, now.Add(-SuppressionWindow),
	).Scan(&out).Error
	return out, err
}
