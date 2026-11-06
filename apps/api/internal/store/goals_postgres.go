package store

import "context"

const goalColumns = "id, user_id, space_id, title, category, done, created_at, updated_at"

type goalRow Goal

func (p *Postgres) goals(ctx context.Context, query string, args ...any) ([]Goal, error) {
	var rows []goalRow
	if err := p.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]Goal, len(rows))
	for i, row := range rows {
		out[i] = Goal(row)
	}
	return out, nil
}

func (p *Postgres) oneGoal(ctx context.Context, query string, args ...any) (Goal, error) {
	found, err := p.goals(ctx, query, args...)
	if err != nil {
		return Goal{}, err
	}
	if len(found) == 0 {
		return Goal{}, ErrNotFound
	}
	return found[0], nil
}

func (p *Postgres) ListGoals(ctx context.Context, userID int64, spaceID string) ([]Goal, error) {
	return p.goals(ctx, "SELECT "+goalColumns+
		" FROM goals WHERE user_id = ? AND space_id = ? ORDER BY done, created_at, id", userID, spaceID)
}

func (p *Postgres) GetGoal(ctx context.Context, userID int64, id string) (Goal, error) {
	return p.oneGoal(ctx, "SELECT "+goalColumns+" FROM goals WHERE id = ? AND user_id = ?", id, userID)
}

func (p *Postgres) CreateGoal(ctx context.Context, goal Goal) (Goal, error) {
	if goal.ID == "" {
		goal.ID = NewID()
	}
	return p.oneGoal(ctx, `INSERT INTO goals (id, user_id, space_id, title, category, done)
		SELECT ?, ?, id, ?, ?, ? FROM spaces WHERE id = ? AND user_id = ? RETURNING `+goalColumns,
		goal.ID, goal.UserID, goal.Title, goal.Category, goal.Done, goal.SpaceID, goal.UserID)
}

func (p *Postgres) UpdateGoal(ctx context.Context, goal Goal) (Goal, error) {
	return p.oneGoal(ctx, `UPDATE goals SET title = ?, category = ?, done = ?, updated_at = now()
		WHERE id = ? AND user_id = ? RETURNING `+goalColumns,
		goal.Title, goal.Category, goal.Done, goal.ID, goal.UserID)
}

func (p *Postgres) DeleteGoal(ctx context.Context, userID int64, id string) error {
	result := p.db.WithContext(ctx).Exec("DELETE FROM goals WHERE id = ? AND user_id = ?", id, userID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
