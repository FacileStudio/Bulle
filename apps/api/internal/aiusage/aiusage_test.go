package aiusage_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/aiusage"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/migrations"
	"github.com/FacileStudio/nacelle"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/migrate"
	"github.com/FacileStudio/tronc/testdb"
	"gorm.io/gorm"
)

var testDB *gorm.DB

var dbConfig = testdb.Config{
	Prefix: "bulle_aiusage_test",
	Migrate: func(db *gorm.DB) error {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return migrate.Run(context.Background(), migrate.Config{
			DB: sqlDB, FS: migrations.FS, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), LockID: 7_300_331,
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

var now = time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC)

func fixture(t *testing.T) *aiusage.Budget {
	t.Helper()
	if testDB == nil {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	exec(t, `DELETE FROM users`)
	exec(t, `INSERT INTO users (id, email) VALUES (1, 'a@x.test'), (2, 'Owner@x.test')`)
	st := store.NewPostgres(testDB)
	for id, user := range map[string]int64{"s1": 1, "s2": 2} {
		if _, err := st.CreateSession(context.Background(), store.Session{ID: id, UserID: user, Channel: store.ChannelWeb}); err != nil {
			t.Fatal(err)
		}
	}
	return &aiusage.Budget{DB: testDB, DefaultUSD: 5, Now: func() time.Time { return now }}
}

func exec(t *testing.T, query string, args ...any) {
	t.Helper()
	if err := testDB.Exec(query, args...).Error; err != nil {
		t.Fatal(err)
	}
}

type spending struct {
	user     int64
	run, app float64
	own      bool
	at       time.Time
}

func spend(t *testing.T, s spending) {
	t.Helper()
	session := map[int64]string{1: "s1", 2: "s2"}[s.user]
	exec(t, `INSERT INTO runs (id, session_id, status, cost_usd, own_key, started_at) VALUES (?, ?, 'finished', ?, ?, ?)`,
		fmt.Sprintf("r%d%v%v", s.user, s.own, s.at.Unix()), session, s.run, s.own, s.at)
	exec(t, `INSERT INTO ai_usage (user_id, source, cost_usd, own_key, created_at) VALUES (?, 'complete', ?, ?, ?)`, s.user, s.app, s.own, s.at)
}

func TestRecordWritesOneRow(t *testing.T) {
	fixture(t)
	call := aiusage.Call{Model: "m", Usage: nacelle.Usage{InputTokens: 11, OutputTokens: 4, Cost: 0.25}}
	aiusage.NewLedger(testDB, slog.New(slog.NewTextHandler(io.Discard, nil))).Record(context.Background(), 1, aiusage.SourceSummarize, "sonar", call)
	var row struct {
		Source, Caller, Model string
		InputTokens           int64
		CostUSD               float64
		OwnKey                bool
	}
	if err := testDB.Raw(`SELECT source, caller, model, input_tokens, cost_usd::float8 AS cost_usd, own_key FROM ai_usage WHERE user_id = 1`).Scan(&row).Error; err != nil {
		t.Fatal(err)
	}
	if row.Source != "summarize" || row.Caller != "sonar" || row.Model != "m" || row.InputTokens != 11 || row.CostUSD != 0.25 || row.OwnKey {
		t.Fatalf("row %+v", row)
	}
}

func TestSpendSumsRunsAndAppCallsOfTheServerKeysThisMonth(t *testing.T) {
	budget := fixture(t)
	thisMonth := now.Add(-time.Hour)
	spend(t, spending{1, 1, 0.5, false, thisMonth})
	spend(t, spending{1, 100, 100, true, thisMonth.Add(time.Minute)})
	spend(t, spending{1, 100, 100, false, time.Date(2026, 10, 31, 23, 0, 0, 0, time.UTC)})
	spend(t, spending{2, 50, 50, false, thisMonth})
	status, err := budget.Status(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if status.SpentUSD != 1.5 || status.LimitUSD != 5 || !status.ResetsAt.Equal(time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("status %+v", status)
	}
}

func TestCheckRefusesOverTheLimitWith429AndTheResetDate(t *testing.T) {
	budget := fixture(t)
	spend(t, spending{1, 3, 2, false, now})
	err := budget.Check(context.Background(), 1)
	if troncerrors.Status(err) != 429 || err.Error() != "This month's AI allowance is used up. It resets on 1 December." {
		t.Fatalf("got %v", err)
	}
	if err := budget.Check(context.Background(), 2); err != nil {
		t.Fatalf("a user under the limit was refused: %v", err)
	}
}

func TestCheckLetsAnExemptEmailThrough(t *testing.T) {
	budget := fixture(t)
	spend(t, spending{2, 30, 0, false, now})
	if err := budget.Check(context.Background(), 2); err == nil {
		t.Fatal("not exempt yet")
	}
	budget.Exempt = []string{"owner@x.test"}
	if err := budget.Check(context.Background(), 2); err != nil {
		t.Fatalf("exempt user refused: %v", err)
	}
}

func TestZeroBudgetDisablesTheCheck(t *testing.T) {
	budget := fixture(t)
	budget.DefaultUSD = 0
	spend(t, spending{1, 1000, 0, false, now})
	if err := budget.Check(context.Background(), 1); err != nil {
		t.Fatalf("budget 0 refused: %v", err)
	}
}

func TestPlanRaisesTheLimit(t *testing.T) {
	budget := fixture(t)
	spend(t, spending{1, 8, 0, false, now})
	budget.Plan = func(context.Context, int64, float64) float64 { return 20 }
	if err := budget.Check(context.Background(), 1); err != nil {
		t.Fatalf("plan limit ignored: %v", err)
	}
}
