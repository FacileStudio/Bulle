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

func TestValidateRefusesWhatTheBotCannotUse(t *testing.T) {
	ok := Settings{Enabled: true, Homeserver: "https://matrix.org", User: "@bot:matrix.org", Allow: []string{"@me:matrix.org"}, Rooms: map[string]string{"s1": "!room:matrix.org"}}
	if err := validate(ok); err != nil {
		t.Fatalf("valid settings refused: %v", err)
	}
	bad := []Settings{
		{Homeserver: "http://matrix.org"},
		{User: "bot"},
		{Allow: []string{"me"}},
		{Rooms: map[string]string{"s1": "#alias:matrix.org"}},
		{Enabled: true},
	}
	for _, in := range bad {
		if err := validate(in); !errors.Is(err, ErrInvalid) {
			t.Errorf("validate(%+v) = %v, want ErrInvalid", in, err)
		}
	}
	if got := clean(Settings{Homeserver: " https://m.org/ ", Allow: []string{" @a:b ", ""}}); got.Homeserver != "https://m.org" || len(got.Allow) != 1 {
		t.Errorf("clean = %+v", got)
	}
}

func TestPostgresRepoLinksSpacesAndThreads(t *testing.T) {
	if testDB == nil {
		t.Skip("no TEST_DATABASE_URL")
	}
	ctx := context.Background()
	for _, q := range []string{
		"TRUNCATE users RESTART IDENTITY CASCADE",
		"INSERT INTO users (id, email) VALUES (1, 'a@facile.studio'), (2, 'b@facile.studio')",
		"INSERT INTO spaces (id, user_id, name) VALUES ('s1', 1, 'Work'), ('s2', 2, 'Other')",
		"INSERT INTO sessions (id, user_id, space_id, channel) VALUES ('sess1', 1, 's1', 'matrix')",
	} {
		if err := testDB.Exec(q).Error; err != nil {
			t.Fatal(err)
		}
	}
	sqlDB, err := testDB.DB()
	if err != nil {
		t.Fatal(err)
	}
	r := repo{db: sqlDB}
	if err := r.put(ctx, Bot{UserID: 1, Enabled: true, Homeserver: "https://m.org", User: "@b:m.org", Password: []byte("p"), Nonce: []byte("n"), Allow: []string{"@me:m.org"}}); err != nil {
		t.Fatal(err)
	}
	if err := r.put(ctx, Bot{UserID: 1, Enabled: true, Homeserver: "https://m.org", User: "@b:m.org"}); err != nil {
		t.Fatal(err)
	}
	bot, found, err := r.get(ctx, 1)
	if err != nil || !found || string(bot.Password) != "p" || len(bot.Allow) != 0 {
		t.Fatalf("get = %+v %v %v, want the password kept on a save without one", bot, found, err)
	}
	if err := r.setSpaceRooms(ctx, 1, map[string]string{"s2": "!x:m.org"}); !errors.Is(err, ErrUnknownSpace) {
		t.Fatalf("linking another user's space: %v", err)
	}
	if err := r.setSpaceRooms(ctx, 1, map[string]string{"s1": "!room:m.org"}); err != nil {
		t.Fatal(err)
	}
	if space, err := r.roomSpace(ctx, 1, "!room:m.org"); err != nil || space != "s1" {
		t.Fatalf("roomSpace = %q %v", space, err)
	}
	if err := r.setThreadSession(ctx, 1, "!room:m.org", "$root", "sess1"); err != nil {
		t.Fatal(err)
	}
	if id, err := r.threadSession(ctx, 1, "!room:m.org", "$root"); err != nil || id != "sess1" {
		t.Fatalf("threadSession = %q %v", id, err)
	}
	if bots, err := r.enabled(ctx); err != nil || len(bots) != 1 {
		t.Fatalf("enabled = %d %v", len(bots), err)
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
	if _, err := db.RawDB.Exec("CREATE TABLE IF NOT EXISTS probe (x int)"); err != nil {
		t.Fatal(err)
	}
	var schema string
	if err := sqlDB.QueryRow("SELECT table_schema FROM information_schema.tables WHERE table_name = 'probe'").Scan(&schema); err != nil || schema != "matrix_u42" {
		t.Fatalf("probe table in %q (%v), want matrix_u42", schema, err)
	}
}
