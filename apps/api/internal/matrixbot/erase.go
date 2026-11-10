package matrixbot

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

// Erase stops the person's bot and drops every crypto store it ever used, for
// an account Registre erased. The rows in matrix_bots and the room tables go
// with the users row.
func (m *Manager) Erase(ctx context.Context, userID int64) error {
	m.halt(userID)
	conf, err := pgx.ParseConfig(m.cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("matrix: database url: %w", err)
	}
	return m.dropStores(ctx, conf.Database, userID, "")
}

// sweepOrphans drops the stores of users that no longer exist: accounts
// erased before their stores were dropped with them.
func (m *Manager) sweepOrphans(ctx context.Context) {
	conf, err := pgx.ParseConfig(m.cfg.DatabaseURL)
	if err != nil {
		return
	}
	ids, err := m.orphanUsers(ctx, conf.Database)
	if err != nil {
		m.cfg.Logger.Warn("matrix: listing orphan crypto stores", slog.Any("error", err))
		return
	}
	for _, id := range ids {
		if err := m.dropStores(ctx, conf.Database, id, ""); err != nil {
			m.cfg.Logger.Warn("matrix: dropping orphan crypto stores", slog.Int64("user_id", id), slog.Any("error", err))
		}
	}
}

// orphanUsers lists the user ids that have a crypto store but no users row.
func (m *Manager) orphanUsers(ctx context.Context, base string) ([]int64, error) {
	rows, err := m.cfg.SQL.QueryContext(ctx, `WITH stores AS (
			SELECT CASE WHEN part ~ '^[0-9]+$' THEN part::bigint END AS id
			FROM (SELECT split_part(substr(datname, length($1) + 1), '_', 1) AS part
				FROM pg_database WHERE starts_with(datname, $1)) AS parts)
		SELECT DISTINCT id FROM stores
		WHERE id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM users WHERE users.id = stores.id)`, cryptoPrefix(base, 0)[:len(base)+4])
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
