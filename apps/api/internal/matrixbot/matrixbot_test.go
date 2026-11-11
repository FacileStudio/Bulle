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
	"github.com/jackc/pgx/v5"
	"go.mau.fi/util/dbutil"
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

func cryptoManager(t *testing.T) *Manager {
	t.Helper()
	url, ok := testdb.URL()
	if !ok || testDB == nil {
		t.Skip("no TEST_DATABASE_URL")
	}
	sqlDB, err := testDB.DB()
	if err != nil {
		t.Fatal(err)
	}
	return &Manager{cfg: Config{SQL: sqlDB, DatabaseURL: url}, erased: map[int64]bool{}}
}

func cryptoBase(t *testing.T, m *Manager) string {
	t.Helper()
	conf, err := pgx.ParseConfig(m.cfg.DatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	return conf.Database
}

func openCrypto(t *testing.T, m *Manager, userID int64, device string) *dbutil.Database {
	t.Helper()
	db, err := m.cryptoDB(context.Background(), userID, device)
	if err != nil {
		t.Fatal(err)
	}
	name := cryptoName(cryptoBase(t, m), userID, device)
	t.Cleanup(func() {
		_ = db.RawDB.Close()
		_, _ = m.cfg.SQL.Exec("DROP DATABASE IF EXISTS " + pgx.Identifier{name}.Sanitize() + " WITH (FORCE)")
	})
	return db
}

func TestCryptoStoresOfTwoBotsNeverSeeEachOthersTables(t *testing.T) {
	m := cryptoManager(t)
	first := openCrypto(t, m, 941, "AAAA")
	second := openCrypto(t, m, 942, "BBBB")
	if got := first.RawDB.Stats().MaxOpenConnections; got != cryptoConns {
		t.Fatalf("crypto pool allows %d connections, want %d", got, cryptoConns)
	}
	if _, err := first.RawDB.Exec("CREATE TABLE probe (x int)"); err != nil {
		t.Fatal(err)
	}
	exists, err := second.TableExists(context.Background(), "probe")
	if err != nil || exists {
		t.Fatalf("second bot sees the first bot's table: exists=%v err=%v", exists, err)
	}
}

func TestDropStaleRemovesEarlierDevicesOnly(t *testing.T) {
	m := cryptoManager(t)
	old := openCrypto(t, m, 943, "OLD")
	_ = old.RawDB.Close()
	openCrypto(t, m, 943, "NEW")
	if err := m.dropStale(context.Background(), 943, "NEW"); err != nil {
		t.Fatal(err)
	}
	base := cryptoBase(t, m)
	for device, want := range map[string]bool{"OLD": false, "NEW": true} {
		var exists bool
		if err := m.cfg.SQL.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", cryptoName(base, 943, device)).Scan(&exists); err != nil {
			t.Fatal(err)
		}
		if exists != want {
			t.Fatalf("store for device %s exists=%v, want %v", device, exists, want)
		}
	}
}
