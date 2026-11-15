package plans_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/plans"
	"github.com/FacileStudio/porte/directory"
)

func TestMonthlyBudgetTakesTheMostGenerousPlan(t *testing.T) {
	if testDB == nil {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	exec(t, `TRUNCATE forfait_state, porte_identities, users RESTART IDENTITY CASCADE`)
	exec(t, `INSERT INTO users (id, email) VALUES (1, 'a@x.test'), (2, 'b@x.test'), (3, 'c@x.test')`)
	exec(t, `INSERT INTO porte_identities (user_id, provider, subject) VALUES (1, ?, 'sub-a'), (2, ?, 'sub-b'), (2, ?, 'sub-b2'), (3, ?, 'sub-c')`,
		issuer, issuer, issuer, issuer)
	ctx := context.Background()
	gate := newGate()
	if got := gate.MonthlyBudget(ctx, 1, 5); got != 5 {
		t.Fatalf("before the first snapshot: %v", got)
	}
	person := func(subject, value string) directory.Person {
		return directory.Person{Subject: subject, Entitlements: map[string]json.RawMessage{plans.AIMonthlyBudget: json.RawMessage(value)}}
	}
	snapshot := directory.Snapshot{Version: 1, EntitlementsRevision: 1, People: []directory.Person{
		person("sub-a", "2"), person("sub-b", "3"), person("sub-b2", "25"), person("sub-c", "-1"),
	}}
	if err := gate.Apply(ctx, snapshot); err != nil {
		t.Fatal(err)
	}
	for user, want := range map[int64]float64{1: 5, 2: 25, 3: 0, 99: 5} {
		if got := gate.MonthlyBudget(ctx, user, 5); got != want {
			t.Errorf("user %d: %v, want %v", user, got, want)
		}
	}
}
