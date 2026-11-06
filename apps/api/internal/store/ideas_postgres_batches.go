package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

const batchColumns = "id, user_id, space_id, run_id, trigger, status, created_at, finished_at"

type batchRow IdeaBatch

func (p *Postgres) batches(ctx context.Context, query string, args ...any) ([]IdeaBatch, error) {
	var rows []batchRow
	if err := p.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]IdeaBatch, len(rows))
	for i, row := range rows {
		out[i] = IdeaBatch(row)
	}
	return out, nil
}

func (p *Postgres) CreateIdeaBatch(ctx context.Context, batch IdeaBatch) (IdeaBatch, error) {
	if batch.ID == "" {
		batch.ID = NewID()
	}
	found, err := p.batches(ctx, `INSERT INTO idea_batches (id, user_id, space_id, run_id, trigger, status)
		SELECT ?, ?, id, ?, ?, 'running' FROM spaces WHERE id = ? AND user_id = ? RETURNING `+batchColumns,
		batch.ID, batch.UserID, batch.RunID, batch.Trigger, batch.SpaceID, batch.UserID)
	if errors.Is(err, gorm.ErrDuplicatedKey) || (err != nil && strings.Contains(err.Error(), "idea_batches_one_running")) {
		return IdeaBatch{}, ErrIdeaBatchRunning
	}
	if err != nil {
		return IdeaBatch{}, err
	}
	if len(found) == 0 {
		return IdeaBatch{}, ErrNotFound
	}
	return found[0], nil
}

func (p *Postgres) UpdateIdeaBatch(ctx context.Context, batch IdeaBatch) error {
	result := p.db.WithContext(ctx).Exec(
		"UPDATE idea_batches SET run_id = ?, status = ?, finished_at = ? WHERE id = ? AND user_id = ?",
		batch.RunID, batch.Status, batch.FinishedAt, batch.ID, batch.UserID,
	)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *Postgres) RunningIdeaBatch(ctx context.Context, userID int64, spaceID string) (IdeaBatch, bool, error) {
	found, err := p.batches(ctx, "SELECT "+batchColumns+
		" FROM idea_batches WHERE user_id = ? AND space_id = ? AND status = 'running' LIMIT 1", userID, spaceID)
	if err != nil || len(found) == 0 {
		return IdeaBatch{}, false, err
	}
	return found[0], true, nil
}

func (p *Postgres) InterruptIdeaBatches(ctx context.Context) error {
	return p.db.WithContext(ctx).Exec(
		"UPDATE idea_batches SET status = 'error', finished_at = now() WHERE status = 'running'",
	).Error
}

func (p *Postgres) LastIdeaBatch(ctx context.Context, userID int64, spaceID string) (IdeaBatch, bool, error) {
	found, err := p.batches(ctx, "SELECT "+batchColumns+` FROM idea_batches
		WHERE user_id = ? AND space_id = ? AND status = 'finished' ORDER BY created_at DESC LIMIT 1`, userID, spaceID)
	if err != nil || len(found) == 0 {
		return IdeaBatch{}, false, err
	}
	return found[0], true, nil
}

func (p *Postgres) CountIdeaBatches(ctx context.Context, userID int64, spaceID, trigger string, since time.Time) (int, error) {
	var count int
	err := p.db.WithContext(ctx).Raw(
		"SELECT count(*) FROM idea_batches WHERE user_id = ? AND space_id = ? AND trigger = ? AND created_at >= ?",
		userID, spaceID, trigger, since,
	).Scan(&count).Error
	return count, err
}
