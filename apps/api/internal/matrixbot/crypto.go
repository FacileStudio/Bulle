package matrixbot

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
	"go.mau.fi/util/dbutil"
)

// cryptoConns caps each bot's crypto pool. runOnce closes the pool when the
// bot stops, because mautrix never closes a database it was handed.
const cryptoConns = 3

// cryptoName is the database holding one bot device's mautrix store. A
// database, not a schema: mautrix checks for its tables in
// information_schema by name alone, so two bots sharing a database see each
// other's tables. A new device gets a fresh store, as mautrix requires.
func cryptoName(base string, userID int64, deviceID string) string {
	device := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, strings.ToLower(deviceID))
	return fmt.Sprintf("%s_mx_%d_%s", base, userID, device)
}

func cryptoPrefix(base string, userID int64) string {
	return fmt.Sprintf("%s_mx_%d_", base, userID)
}

// cryptoDB opens the bot device's own database, creating it on first use.
func (m *Manager) cryptoDB(ctx context.Context, userID int64, deviceID string) (*dbutil.Database, error) {
	conf, err := pgx.ParseConfig(m.cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("matrix: database url: %w", err)
	}
	name := cryptoName(conf.Database, userID, deviceID)
	if err := m.ensureDatabase(ctx, name); err != nil {
		return nil, err
	}
	conf.Database = name
	raw := stdlib.OpenDB(*conf)
	raw.SetMaxOpenConns(cryptoConns)
	raw.SetMaxIdleConns(1)
	db, err := dbutil.NewWithDB(raw, "postgres")
	if err != nil {
		_ = raw.Close()
		return nil, fmt.Errorf("matrix: opening %s: %w", name, err)
	}
	return db, nil
}

func (m *Manager) ensureDatabase(ctx context.Context, name string) error {
	var exists bool
	if err := m.cfg.SQL.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", name).Scan(&exists); err != nil {
		return fmt.Errorf("matrix: looking up %s: %w", name, err)
	}
	if exists {
		return nil
	}
	_, err := m.cfg.SQL.ExecContext(ctx, "CREATE DATABASE "+pgx.Identifier{name}.Sanitize())
	var pgErr *pgconn.PgError
	if err != nil && (!errors.As(err, &pgErr) || pgErr.Code != "42P04") {
		return fmt.Errorf("matrix: creating %s: %w", name, err)
	}
	return nil
}

// dropStale removes the stores of the bot's earlier devices and the schema
// the bot used before stores moved to their own databases.
func (m *Manager) dropStale(ctx context.Context, userID int64, deviceID string) error {
	conf, err := pgx.ParseConfig(m.cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("matrix: database url: %w", err)
	}
	keep := cryptoName(conf.Database, userID, deviceID)
	rows, err := m.cfg.SQL.QueryContext(ctx, "SELECT datname FROM pg_database WHERE starts_with(datname, $1) AND datname <> $2", cryptoPrefix(conf.Database, userID), keep)
	if err != nil {
		return fmt.Errorf("matrix: listing stores: %w", err)
	}
	var stale []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			_ = rows.Close()
			return fmt.Errorf("matrix: listing stores: %w", err)
		}
		stale = append(stale, name)
	}
	_ = rows.Close()
	failures := []error{rows.Err()}
	for _, name := range stale {
		_, err := m.cfg.SQL.ExecContext(ctx, "DROP DATABASE IF EXISTS "+pgx.Identifier{name}.Sanitize()+" WITH (FORCE)")
		failures = append(failures, err)
	}
	_, err = m.cfg.SQL.ExecContext(ctx, fmt.Sprintf("DROP SCHEMA IF EXISTS matrix_u%d CASCADE", userID))
	return errors.Join(append(failures, err)...)
}
