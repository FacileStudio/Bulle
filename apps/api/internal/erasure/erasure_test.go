package erasure

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/migrations"
	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/tronc/migrate"
	"github.com/FacileStudio/tronc/testdb"
	"gorm.io/gorm"
)

const issuer = "https://sso.example.test"

var testDB *gorm.DB

var dbConfig = testdb.Config{
	Prefix: "bulle_erasure_test",
	Migrate: func(db *gorm.DB) error {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return migrate.Run(context.Background(), migrate.Config{
			DB: sqlDB, FS: migrations.FS, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), LockID: 7_300_322,
		})
	},
}

func TestMain(m *testing.M) {
	if url, ok := testdb.URL(); ok {
		db, err := testdb.Open(url, dbConfig)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		testDB = db
	}
	os.Exit(m.Run())
}

func exec(t *testing.T, sql string, args ...any) {
	t.Helper()
	if err := testDB.Exec(sql, args...).Error; err != nil {
		t.Fatal(err)
	}
}

func count(t *testing.T, query string, args ...any) int64 {
	t.Helper()
	var n int64
	if err := testDB.Raw(query, args...).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

func TestErasePersonDeletesOnlyThatPersonAndRepeats(t *testing.T) {
	if testDB == nil {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	if err := testdb.Truncate(testDB, dbConfig); err != nil {
		t.Fatal(err)
	}
	exec(t, `INSERT INTO users (id, email) VALUES (1, 'gone@x.test'), (2, 'kept@x.test')`)
	exec(t, `INSERT INTO porte_identities (user_id, provider, subject) VALUES (1, ?, 'sub-gone'), (2, ?, 'sub-kept')`, issuer, issuer)
	exec(t, `INSERT INTO spaces (id, user_id, name) VALUES ('s1', 1, 'Personal'), ('s2', 2, 'Personal')`)
	exec(t, `INSERT INTO sessions (id, user_id, space_id) VALUES ('c1', 1, 's1'), ('c2', 2, 's2')`)
	exec(t, `INSERT INTO messages (session_id, seq, payload) VALUES ('c1', 0, '{}'), ('c2', 0, '{}')`)
	exec(t, `INSERT INTO jobs (id, user_id, space_id, name, schedule, prompt) VALUES ('j1', '1', 's1', 'n', '* * * * *', 'p'), ('j2', '2', 's2', 'n', '* * * * *', 'p')`)

	var erasedComputers, halted, forgotten []int64
	deletions := New(testDB, directory.Client{Issuer: issuer}, nil, Hooks{
		EraseComputer: func(_ context.Context, id int64) error { erasedComputers = append(erasedComputers, id); return nil },
		HaltBot:       func(id int64) { halted = append(halted, id) },
		ForgetTokens:  func(id int64) { forgotten = append(forgotten, id) },
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))

	for range 2 {
		if err := deletions.ErasePerson(context.Background(), issuer, "sub-gone"); err != nil {
			t.Fatal(err)
		}
	}

	for query, want := range map[string]int64{
		`SELECT count(*) FROM users WHERE id = 1`:                 0,
		`SELECT count(*) FROM porte_identities WHERE user_id = 1`: 0,
		`SELECT count(*) FROM spaces WHERE user_id = 1`:           0,
		`SELECT count(*) FROM sessions WHERE user_id = 1`:         0,
		`SELECT count(*) FROM messages WHERE session_id = 'c1'`:   0,
		`SELECT count(*) FROM jobs WHERE id = 'j1'`:               0,
		`SELECT count(*) FROM users WHERE id = 2`:                 1,
		`SELECT count(*) FROM porte_identities WHERE user_id = 2`: 1,
		`SELECT count(*) FROM spaces WHERE user_id = 2`:           1,
		`SELECT count(*) FROM messages WHERE session_id = 'c2'`:   1,
		`SELECT count(*) FROM jobs WHERE id = 'j2'`:               1,
	} {
		if got := count(t, query); got != want {
			t.Errorf("%s = %d, want %d", query, got, want)
		}
	}
	if len(erasedComputers) != 1 || len(halted) != 1 || len(forgotten) != 1 || erasedComputers[0] != 1 {
		t.Errorf("hooks ran for %v %v %v, want user 1 once each", erasedComputers, halted, forgotten)
	}
}

func TestStateSurvivesARestart(t *testing.T) {
	if testDB == nil {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	if err := testdb.Truncate(testDB, dbConfig); err != nil {
		t.Fatal(err)
	}
	saved := state{orm: testDB}
	if loaded, err := saved.Load(context.Background()); err != nil || loaded != nil {
		t.Fatalf("empty Load = %q, %v; want nil", loaded, err)
	}
	for _, blob := range []string{"one", "two"} {
		if err := saved.Save(context.Background(), []byte(blob)); err != nil {
			t.Fatal(err)
		}
	}
	if loaded, err := saved.Load(context.Background()); err != nil || string(loaded) != "two" {
		t.Fatalf("Load = %q, %v; want two", loaded, err)
	}
}
