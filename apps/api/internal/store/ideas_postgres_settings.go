package store

import (
	"context"
)

const ideaSettingsColumns = "space_id, user_id, enabled, hour, timezone, push, steering, last_batch_date, updated_at"

type ideaSettingsRow IdeaSettings

func (p *Postgres) settings(ctx context.Context, query string, args ...any) ([]IdeaSettings, error) {
	var rows []ideaSettingsRow
	if err := p.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]IdeaSettings, len(rows))
	for i, row := range rows {
		out[i] = IdeaSettings(row)
	}
	return out, nil
}

func (p *Postgres) GetIdeaSettings(ctx context.Context, userID int64, spaceID string) (IdeaSettings, error) {
	found, err := p.settings(ctx, "SELECT "+ideaSettingsColumns+
		" FROM idea_settings WHERE space_id = ? AND user_id = ?", spaceID, userID)
	if err != nil {
		return IdeaSettings{}, err
	}
	if len(found) == 0 {
		return IdeaSettings{}, ErrNotFound
	}
	return found[0], nil
}

func (p *Postgres) SaveIdeaSettings(ctx context.Context, s IdeaSettings) (IdeaSettings, error) {
	found, err := p.settings(ctx, `INSERT INTO idea_settings (space_id, user_id, enabled, hour, timezone, push, steering)
		SELECT id, user_id, ?, ?, ?, ?, ? FROM spaces WHERE id = ? AND user_id = ?
		ON CONFLICT (space_id) DO UPDATE SET enabled = EXCLUDED.enabled, hour = EXCLUDED.hour,
			timezone = EXCLUDED.timezone, push = EXCLUDED.push, steering = EXCLUDED.steering, updated_at = now()
		WHERE idea_settings.user_id = EXCLUDED.user_id
		RETURNING `+ideaSettingsColumns,
		s.Enabled, s.Hour, s.Timezone, s.Push, s.Steering, s.SpaceID, s.UserID)
	if err != nil {
		return IdeaSettings{}, err
	}
	if len(found) == 0 {
		return IdeaSettings{}, ErrNotFound
	}
	return found[0], nil
}

func (p *Postgres) ListEnabledIdeaSettings(ctx context.Context) ([]IdeaSettings, error) {
	return p.settings(ctx, "SELECT "+ideaSettingsColumns+" FROM idea_settings WHERE enabled ORDER BY space_id")
}

func (p *Postgres) MarkIdeaBatchDate(ctx context.Context, userID int64, spaceID, date string) (bool, error) {
	result := p.db.WithContext(ctx).Exec(
		"UPDATE idea_settings SET last_batch_date = ? WHERE space_id = ? AND user_id = ? AND last_batch_date <> ?",
		date, spaceID, userID, date,
	)
	return result.RowsAffected > 0, result.Error
}
