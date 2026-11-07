package plans_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/plans"
	"github.com/FacileStudio/Bulle/apps/api/migrations"
	"github.com/FacileStudio/porte/directory"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/migrate"
	"github.com/FacileStudio/tronc/testdb"
	"gorm.io/gorm"
)

const issuer = "https://sso.example.test"

var testDB *gorm.DB

func TestMain(m *testing.M) {
	if url, ok := testdb.URL(); ok {
		db, err := testdb.Open(url, testdb.Config{
			Prefix: "bulle_plans_test",
			Migrate: func(db *gorm.DB) error {
				sqlDB, err := db.DB()
				if err != nil {
					return err
				}
				return migrate.Run(context.Background(), migrate.Config{
					DB: sqlDB, FS: migrations.FS, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), LockID: 7_300_313,
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

func exec(t *testing.T, sql string, args ...any) {
	t.Helper()
	if err := testDB.Exec(sql, args...).Error; err != nil {
		t.Fatal(err)
	}
}

func newGate() *plans.Gate {
	return plans.New(testDB, issuer, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func snapshot(revision int64, people map[string]bool) directory.Snapshot {
	out := directory.Snapshot{Version: revision, EntitlementsRevision: revision}
	for sub, on := range people {
		out.People = append(out.People, directory.Person{
			Subject:      sub,
			Entitlements: map[string]json.RawMessage{plans.AIAccess: json.RawMessage(fmt.Sprint(on))},
		})
	}
	return out
}

func TestCheckAI(t *testing.T) {
	if testDB == nil {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	exec(t, `TRUNCATE forfait_state, porte_identities, users RESTART IDENTITY CASCADE`)
	exec(t, `INSERT INTO users (id, email) VALUES (1, 'free@x.test'), (2, 'paid@x.test'), (3, 'stale@x.test'), (4, 'anon@x.test')`)
	exec(t, `INSERT INTO porte_identities (user_id, provider, subject) VALUES
		(1, ?, 'sub-free'), (2, ?, 'sub-paid'), (3, ?, 'sub-old'), (3, ?, 'sub-new'), (4, 'https://other.test', 'sub-anon')`,
		issuer, issuer, issuer, issuer)
	ctx := context.Background()
	gate := newGate()
	if err := gate.Load(ctx); err != nil {
		t.Fatal(err)
	}

	for user := int64(1); user <= 4; user++ {
		if err := gate.CheckAI(ctx, user); err != nil {
			t.Fatalf("user %d before the first snapshot: %v", user, err)
		}
	}

	snap := snapshot(1, map[string]bool{"sub-free": false, "sub-paid": true, "sub-new": true, "sub-old": false})
	if err := gate.Apply(ctx, snap); err != nil {
		t.Fatal(err)
	}
	if !gate.Active() {
		t.Fatal("the gate must be authoritative after a revision above zero")
	}

	if err := gate.CheckAI(ctx, 2); err != nil {
		t.Fatalf("paid user refused: %v", err)
	}
	if err := gate.CheckAI(ctx, 3); err != nil {
		t.Fatalf("a user with one paid identity must pass: %v", err)
	}
	for _, user := range []int64{1, 4, 99} {
		err := gate.CheckAI(ctx, user)
		if !errors.Is(err, plans.ErrAIAccess) {
			t.Fatalf("user %d must be refused, got %v", user, err)
		}
		var envelope *troncerrors.Error
		if !errors.As(err, &envelope) || envelope.Code != "payment_required" || envelope.UpgradeURL != issuer+"/spaces" {
			t.Fatalf("user %d: want payment_required with the upgrade url, got %+v", user, envelope)
		}
		if troncerrors.Status(envelope) != 402 {
			t.Fatalf("status %d", troncerrors.Status(envelope))
		}
	}

	restarted := newGate()
	if err := restarted.Load(ctx); err != nil {
		t.Fatal(err)
	}
	if !restarted.Active() {
		t.Fatal("a restart must not fail open again")
	}
	if err := restarted.CheckAI(ctx, 1); !errors.Is(err, plans.ErrAIAccess) {
		t.Fatalf("restored gate must keep refusing: %v", err)
	}
	if err := restarted.CheckAI(ctx, 2); err != nil {
		t.Fatalf("restored gate must keep allowing: %v", err)
	}
}

func TestNilGateAllows(t *testing.T) {
	var gate *plans.Gate
	if err := gate.CheckAI(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
}
