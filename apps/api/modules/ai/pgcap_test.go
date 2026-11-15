package ai

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/migrations"
	"github.com/FacileStudio/tronc/migrate"
	"github.com/FacileStudio/tronc/testdb"
	"gorm.io/gorm"
)

var testDB *gorm.DB

func migrateTestDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return migrate.Run(context.Background(), migrate.Config{
		DB: sqlDB, FS: migrations.FS, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), LockID: 7_300_313,
	})
}

func TestMain(m *testing.M) {
	if url, ok := testdb.URL(); ok {
		db, err := testdb.Open(url, testdb.Config{Prefix: "bulle_ai_test", Migrate: migrateTestDB})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		testDB = db
	}
	os.Exit(m.Run())
}

func resetCapRows(t *testing.T) {
	t.Helper()
	for _, q := range []string{
		`INSERT INTO users (id, email) VALUES (901, 'cap@x.test'), (902, 'other@x.test') ON CONFLICT DO NOTHING`,
		`DELETE FROM ai_daily_usage`,
	} {
		if err := testDB.Exec(q).Error; err != nil {
			t.Fatal(err)
		}
	}
}

func TestPgCapSurvivesRestartAndRefusesPastLimit(t *testing.T) {
	if testDB == nil {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	resetCapRows(t)
	now := time.Date(2026, 11, 11, 23, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	first := &pgCap{db: testDB, limit: 3, now: clock}
	for i := 0; i < 2; i++ {
		if !take(first, 901) {
			t.Fatalf("completion %d refused", i+1)
		}
	}
	restarted := &pgCap{db: testDB, limit: 3, now: clock}
	if !take(restarted, 901) {
		t.Fatal("third completion refused")
	}
	if take(restarted, 901) {
		t.Fatal("completion past the limit was allowed after a restart")
	}
	if !take(restarted, 902) {
		t.Fatal("one user's cap blocked another")
	}
	now = now.Add(2 * time.Hour)
	if !take(restarted, 901) {
		t.Fatal("the cap did not reset the next UTC day")
	}
}
