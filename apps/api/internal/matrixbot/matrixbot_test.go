package matrixbot

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/migrations"
	"github.com/FacileStudio/tronc/migrate"
	"github.com/FacileStudio/tronc/testdb"
	"gorm.io/gorm"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	if url, ok := testdb.URL(); ok {
		db, err := testdb.Open(url, testdb.Config{
			Prefix: "bulle_matrix_test",
			Migrate: func(db *gorm.DB) error {
				sqlDB, err := db.DB()
				if err != nil {
					return err
				}
				return migrate.Run(context.Background(), migrate.Config{
					DB: sqlDB, FS: migrations.FS, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), LockID: 7_300_311,
				})
			},
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		testDB = db
	}
	os.Exit(m.Run())
}

func TestSealRoundTripsAndBindsTheUser(t *testing.T) {
	s, err := newSealer(bytes.Repeat([]byte{7}, 32))
	if err != nil {
		t.Fatal(err)
	}
	sealed, nonce, err := s.seal(1, "hunter2")
	if err != nil {
		t.Fatal(err)
	}
	if got, err := s.open(1, sealed, nonce); err != nil || got != "hunter2" {
		t.Fatalf("open = %q, %v", got, err)
	}
	if _, err := s.open(2, sealed, nonce); err == nil {
		t.Fatal("another user's id opened the password")
	}
	if bytes.Equal(s.pickleKey(1), s.pickleKey(2)) {
		t.Fatal("two bots share a pickle key")
	}
	var none *sealer
	if _, _, err := none.seal(1, "x"); !errors.Is(err, ErrDisabled) {
		t.Fatalf("no key: %v, want ErrDisabled", err)
	}
}

func TestCryptoDBUsesTheBotsOwnSchema(t *testing.T) {
	url, ok := testdb.URL()
	if !ok || testDB == nil {
		t.Skip("no TEST_DATABASE_URL")
	}
	sqlDB, err := testDB.DB()
	if err != nil {
		t.Fatal(err)
	}
	m := &Manager{cfg: Config{SQL: sqlDB, DatabaseURL: url}}
	db, err := m.cryptoDB(context.Background(), 42)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.RawDB.Close()
		_, _ = sqlDB.Exec("DROP SCHEMA matrix_u42 CASCADE")
	}()
	if got := db.RawDB.Stats().MaxOpenConnections; got != cryptoConns {
		t.Fatalf("crypto pool allows %d connections, want %d", got, cryptoConns)
	}
	if _, err := db.RawDB.Exec("CREATE TABLE IF NOT EXISTS probe (x int)"); err != nil {
		t.Fatal(err)
	}
	var schema string
	if err := sqlDB.QueryRow("SELECT table_schema FROM information_schema.tables WHERE table_name = 'probe'").Scan(&schema); err != nil || schema != "matrix_u42" {
		t.Fatalf("probe table in %q (%v), want matrix_u42", schema, err)
	}
}
