package store

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

const chatTextMatch = `EXISTS (SELECT 1 FROM messages m, jsonb_array_elements(m.payload->'parts') part
	WHERE m.session_id = s.id AND m.payload->>'role' IN ('user', 'assistant')
	AND part->>'type' = 'text' AND part->>'text' ILIKE ?)`

func (p *Postgres) ListSessionsForMemoryPass(ctx context.Context, idleSince time.Time) ([]Session, error) {
	var rows []sessionRow
	err := p.db.WithContext(ctx).Raw(
		`SELECT `+prefixed("s.", sessionColumns)+` FROM sessions s
		WHERE s.channel IN (?, ?) AND s.updated_at < ? AND s.updated_at > ?
		AND (SELECT COALESCE(MAX(m.seq) + 1, 0) FROM messages m WHERE m.session_id = s.id) > s.memory_seen_seq
		ORDER BY s.updated_at DESC, s.id LIMIT ?`,
		ChannelWeb, ChannelMatrix, idleSince, idleSince.Add(-MemoryPassLookback), memoryPassBatch,
	).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]Session, len(rows))
	for i, row := range rows {
		out[i] = row.session()
	}
	return out, nil
}

func (p *Postgres) MarkMemorySeen(ctx context.Context, sessionID string, seq int) error {
	result := p.db.WithContext(ctx).Exec(
		"UPDATE sessions SET memory_seen_seq = GREATEST(memory_seen_seq, ?) WHERE id = ?", seq, sessionID,
	)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

type hitRow struct {
	ID        string
	Title     string
	UpdatedAt time.Time
}

// MarkTainted records that a run of the session read a third party's text,
// so its memory passes keep only the owner's own words.
func (p *Postgres) MarkTainted(ctx context.Context, sessionID string) error {
	result := p.db.WithContext(ctx).Exec("UPDATE sessions SET tainted = true WHERE id = ?", sessionID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *Postgres) SearchSessions(ctx context.Context, userID int64, spaceID, query string, limit int) ([]SessionHit, error) {
	query = strings.TrimSpace(query)
	pattern := likePattern(query)
	var rows []hitRow
	err := p.db.WithContext(ctx).Raw(
		`SELECT s.id, s.title, s.updated_at FROM sessions s
		WHERE s.user_id = ? AND s.space_id = ? AND s.channel IN (?, ?)
		AND (s.title ILIKE ? OR `+chatTextMatch+`)
		ORDER BY s.updated_at DESC, s.id LIMIT ?`,
		userID, spaceID, ChannelWeb, ChannelMatrix, pattern, pattern, max(limit, 0),
	).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]SessionHit, len(rows))
	for i, row := range rows {
		snippet, err := p.snippet(ctx, row.ID, query, pattern)
		if err != nil {
			return nil, err
		}
		out[i] = SessionHit{SessionID: row.ID, Title: row.Title, Snippet: snippet, UpdatedAt: row.UpdatedAt}
	}
	return out, nil
}

func (p *Postgres) snippet(ctx context.Context, sessionID, query, pattern string) (string, error) {
	var rows []messageRow
	err := p.db.WithContext(ctx).Raw(
		`SELECT m.seq, m.payload FROM messages m WHERE m.session_id = ? AND m.payload->>'role' IN ('user', 'assistant')
		AND EXISTS (SELECT 1 FROM jsonb_array_elements(m.payload->'parts') part WHERE part->>'type' = 'text' AND part->>'text' ILIKE ?)
		ORDER BY m.seq LIMIT 1`,
		sessionID, pattern,
	).Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return "", err
	}
	snippet, _ := FirstSnippet([]json.RawMessage{rows[0].Payload}, query)
	return snippet, nil
}

func prefixed(prefix, columns string) string {
	names := strings.Split(columns, ", ")
	for i, name := range names {
		names[i] = prefix + name
	}
	return strings.Join(names, ", ")
}
