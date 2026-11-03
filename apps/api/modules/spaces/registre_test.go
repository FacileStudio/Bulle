package spaces_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/migrations"
	"github.com/FacileStudio/Bulle/apps/api/modules/spaces"
	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/tronc/migrate"
	"github.com/FacileStudio/tronc/testdb"
	"gorm.io/gorm"
)

const issuer = "https://sso.example.test"

var testDB *gorm.DB

func TestMain(m *testing.M) {
	if url, ok := testdb.URL(); ok {
		db, err := testdb.Open(url, testdb.Config{
			Prefix: "bulle_spaces_test",
			Migrate: func(db *gorm.DB) error {
				sqlDB, err := db.DB()
				if err != nil {
					return err
				}
				return migrate.Run(context.Background(), migrate.Config{
					DB: sqlDB, FS: migrations.FS, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), LockID: 7_300_312,
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

type spaceRow struct {
	Email           string
	Name            string
	RegistreSpaceID string
}

func spacesOf(t *testing.T) map[string][]spaceRow {
	t.Helper()
	var rows []spaceRow
	err := testDB.Raw(`SELECT u.email, s.name, s.registre_space_id FROM spaces s JOIN users u ON u.id = s.user_id
		ORDER BY u.email, s.created_at, s.id`).Scan(&rows).Error
	if err != nil {
		t.Fatal(err)
	}
	out := map[string][]spaceRow{}
	for _, row := range rows {
		out[row.Email] = append(out[row.Email], row)
	}
	return out
}

func exec(t *testing.T, sql string, args ...any) {
	t.Helper()
	if err := testDB.Exec(sql, args...).Error; err != nil {
		t.Fatal(err)
	}
}

func TestApplyMirrorsSharedSpacesPerMember(t *testing.T) {
	if testDB == nil {
		t.Skip("no TEST_DATABASE_URL")
	}
	exec(t, `TRUNCATE users CASCADE`)
	exec(t, `INSERT INTO users (id, email) VALUES (1, 'alice@x.test'), (2, 'bob@x.test')`)
	exec(t, `SELECT setval(pg_get_serial_sequence('users', 'id'), 2)`)
	exec(t, `INSERT INTO porte_identities (user_id, provider, subject) VALUES (1, ?, 'sub-alice')`, issuer)
	exec(t, `INSERT INTO spaces (id, user_id, name, created_at) VALUES ('a-personal', 1, 'Personal', now() - interval '1 day'),
		('a-team', 1, 'team', now()), ('b-personal', 2, 'Personal', now())`)
	exec(t, `INSERT INTO sessions (id, user_id, space_id, title) VALUES ('chat-1', 1, 'a-team', 'kept')`)

	registre := spaces.NewRegistre(testDB, issuer, slog.New(slog.NewTextHandler(io.Discard, nil)))
	member := func(sub, email string) directory.Member {
		return directory.Member{Subject: sub, Email: email, Role: directory.RoleMember}
	}
	snapshot := directory.Snapshot{Version: 1, Spaces: []directory.Space{
		{ID: "r-team", Name: "Team", Members: []directory.Member{member("sub-alice", "alice@x.test"), member("sub-bob", "BOB@x.test")}},
		{ID: "r-new", Name: "Newcomers", Members: []directory.Member{member("sub-carol", "carol@x.test")}},
	}}
	ctx := context.Background()
	for range 2 {
		if err := registre.Apply(ctx, snapshot); err != nil {
			t.Fatal(err)
		}
	}
	got := spacesOf(t)
	want := map[string][]spaceRow{
		"alice@x.test": {{"alice@x.test", "Personal", ""}, {"alice@x.test", "Team", "r-team"}},
		"bob@x.test":   {{"bob@x.test", "Personal", ""}, {"bob@x.test", "Team", "r-team"}},
		"carol@x.test": {{"carol@x.test", "Personal", ""}, {"carol@x.test", "Newcomers", "r-new"}},
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("after apply\n got %v\nwant %v", got, want)
	}

	snapshot.Spaces[0].Members = snapshot.Spaces[0].Members[1:]
	snapshot.Version = 2
	if err := registre.Apply(ctx, snapshot); err != nil {
		t.Fatal(err)
	}
	if alice := spacesOf(t)["alice@x.test"]; len(alice) != 1 || alice[0].Name != "Personal" {
		t.Fatalf("alice should keep only Personal after leaving, got %v", alice)
	}
	var space string
	if err := testDB.Raw(`SELECT space_id FROM sessions WHERE id = 'chat-1'`).Scan(&space).Error; err != nil || space != "a-personal" {
		t.Fatalf("the chat should move to Personal, got %q (%v)", space, err)
	}
}
