package store

import (
	"context"

	"gorm.io/gorm"
)

const spaceColumns = "id, user_id, name, icon, mycelium_space_id, opus_workspace, agenda_calendar, courrier_account, created_at, updated_at"

type spaceRow Space

func (p *Postgres) ListSpaces(ctx context.Context, userID int64) ([]Space, error) {
	if _, err := p.EnsureDefaultSpace(ctx, userID); err != nil {
		return nil, err
	}
	var rows []spaceRow
	err := p.db.WithContext(ctx).Raw(
		"SELECT "+spaceColumns+" FROM spaces WHERE user_id = ? ORDER BY created_at, id", userID,
	).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]Space, len(rows))
	for i, row := range rows {
		out[i] = Space(row)
	}
	return out, nil
}

func (p *Postgres) GetSpace(ctx context.Context, userID int64, id string) (Space, error) {
	var rows []spaceRow
	err := p.db.WithContext(ctx).Raw(
		"SELECT "+spaceColumns+" FROM spaces WHERE id = ? AND user_id = ?", id, userID,
	).Scan(&rows).Error
	if err != nil {
		return Space{}, err
	}
	if len(rows) == 0 {
		return Space{}, ErrNotFound
	}
	return Space(rows[0]), nil
}

func (p *Postgres) CreateSpace(ctx context.Context, space Space) (Space, error) {
	if space.ID == "" {
		space.ID = NewID()
	}
	var row spaceRow
	err := p.db.WithContext(ctx).Raw(
		`INSERT INTO spaces (id, user_id, name, icon, mycelium_space_id, opus_workspace, agenda_calendar, courrier_account)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?) RETURNING `+spaceColumns,
		space.ID, space.UserID, space.Name, space.Icon, space.MyceliumSpaceID, space.OpusWorkspace,
		space.AgendaCalendar, space.CourrierAccount,
	).Scan(&row).Error
	return Space(row), err
}

func (p *Postgres) UpdateSpace(ctx context.Context, space Space) (Space, error) {
	var rows []spaceRow
	err := p.db.WithContext(ctx).Raw(
		`UPDATE spaces SET name = ?, icon = ?, mycelium_space_id = ?, opus_workspace = ?, agenda_calendar = ?,
		courrier_account = ?, updated_at = now() WHERE id = ? AND user_id = ? RETURNING `+spaceColumns,
		space.Name, space.Icon, space.MyceliumSpaceID, space.OpusWorkspace, space.AgendaCalendar,
		space.CourrierAccount, space.ID, space.UserID,
	).Scan(&rows).Error
	if err != nil {
		return Space{}, err
	}
	if len(rows) == 0 {
		return Space{}, ErrNotFound
	}
	return Space(rows[0]), nil
}

// DeleteSpace locks the user's row first, so two deletes racing for a user's
// last two spaces cannot both pass the count.
func (p *Postgres) DeleteSpace(ctx context.Context, userID int64, id string) error {
	return p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockUser(tx, userID); err != nil {
			return err
		}
		var counts struct {
			Total int
			Found int
		}
		err := tx.Raw(
			"SELECT count(*) AS total, count(*) FILTER (WHERE id = ?) AS found FROM spaces WHERE user_id = ?", id, userID,
		).Scan(&counts).Error
		switch {
		case err != nil:
			return err
		case counts.Found == 0:
			return ErrNotFound
		case counts.Total <= 1:
			return ErrLastSpace
		}
		return tx.Exec("DELETE FROM spaces WHERE id = ? AND user_id = ?", id, userID).Error
	})
}

// EnsureDefaultSpace answers the user's oldest space, creating the default
// one under the user's row lock when there is none yet.
func (p *Postgres) EnsureDefaultSpace(ctx context.Context, userID int64) (Space, error) {
	var space Space
	err := p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockUser(tx, userID); err != nil {
			return err
		}
		var rows []spaceRow
		err := tx.Raw(
			`WITH created AS (
				INSERT INTO spaces (id, user_id, name) SELECT ?, ?, ?
				WHERE NOT EXISTS (SELECT 1 FROM spaces WHERE user_id = ?)
				RETURNING `+spaceColumns+`
			)
			SELECT `+spaceColumns+` FROM created
			UNION ALL
			(SELECT `+spaceColumns+` FROM spaces WHERE user_id = ? ORDER BY created_at, id LIMIT 1)`,
			NewID(), userID, DefaultSpaceName, userID, userID,
		).Scan(&rows).Error
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			return ErrNotFound
		}
		space = Space(rows[0])
		return nil
	})
	return space, err
}

func lockUser(tx *gorm.DB, userID int64) error {
	var found []int64
	if err := tx.Raw("SELECT id FROM users WHERE id = ? FOR UPDATE", userID).Scan(&found).Error; err != nil {
		return err
	}
	if len(found) == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *Postgres) spaceFor(ctx context.Context, userID int64, spaceID string) (string, error) {
	if spaceID != "" {
		return spaceID, nil
	}
	space, err := p.EnsureDefaultSpace(ctx, userID)
	return space.ID, err
}
