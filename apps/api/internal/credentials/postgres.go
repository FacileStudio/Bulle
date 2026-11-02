package credentials

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type postgres struct {
	db *gorm.DB
}

type dbRow struct {
	Backend    string
	Ciphertext []byte
	Nonce      []byte
	Last4      string
	UpdatedAt  time.Time
}

func (d dbRow) row() row {
	return row{Ciphertext: d.Ciphertext, Nonce: d.Nonce, Last4: d.Last4, UpdatedAt: d.UpdatedAt}
}

func NewPostgres(db *gorm.DB, masterKey []byte) (*Service, error) {
	return newService(&postgres{db: db}, masterKey)
}

func (p *postgres) upsert(ctx context.Context, userID int64, backend string, r row) error {
	return p.db.WithContext(ctx).Exec(
		`INSERT INTO provider_credentials (user_id, backend, ciphertext, nonce, last4)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT (user_id, backend) DO UPDATE
		 SET ciphertext = EXCLUDED.ciphertext, nonce = EXCLUDED.nonce, last4 = EXCLUDED.last4, updated_at = now()`,
		userID, backend, r.Ciphertext, r.Nonce, r.Last4,
	).Error
}

func (p *postgres) remove(ctx context.Context, userID int64, backend string) error {
	return p.db.WithContext(ctx).Exec(
		"DELETE FROM provider_credentials WHERE user_id = ? AND backend = ?", userID, backend,
	).Error
}

func (p *postgres) list(ctx context.Context, userID int64) (map[string]row, error) {
	var rows []dbRow
	err := p.db.WithContext(ctx).Raw(
		"SELECT backend, ciphertext, nonce, last4, updated_at FROM provider_credentials WHERE user_id = ?", userID,
	).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]row, len(rows))
	for _, r := range rows {
		out[r.Backend] = r.row()
	}
	return out, nil
}

func (p *postgres) get(ctx context.Context, userID int64, backend string) (row, bool, error) {
	var rows []dbRow
	err := p.db.WithContext(ctx).Raw(
		"SELECT backend, ciphertext, nonce, last4, updated_at FROM provider_credentials WHERE user_id = ? AND backend = ?",
		userID, backend,
	).Scan(&rows).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return row{}, false, err
	}
	if len(rows) == 0 {
		return row{}, false, nil
	}
	return rows[0].row(), true, nil
}
