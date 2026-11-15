package runner_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/aiusage"
	"github.com/FacileStudio/Bulle/apps/api/internal/credentials"
	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/nacelle"
)

type usedUp struct{ checked int }

func (u *usedUp) Check(context.Context, int64) error {
	u.checked++
	return aiusage.ErrBudget
}

func TestStartRefusedWhenTheMonthlyBudgetIsUsedUp(t *testing.T) {
	budget := &usedUp{}
	f := newFixture(t, &fakeBackend{}, nil, func(c *runner.Config) { c.Budget = budget })
	_, err := f.runner.Start(context.Background(), 1, f.session.ID, "hello")
	if !errors.Is(err, aiusage.ErrBudget) {
		t.Fatalf("got %v", err)
	}
	if _, found, _ := f.store.LatestRun(context.Background(), f.session.ID); found {
		t.Fatal("a refused start must not create a run")
	}
}

func TestCompleteRefusedWhenTheMonthlyBudgetIsUsedUp(t *testing.T) {
	f := newFixture(t, &fakeBackend{}, nil, func(c *runner.Config) { c.Budget = &usedUp{} })
	_, _, err := f.runner.Complete(context.Background(), 1, "", "hi")
	if !errors.Is(err, aiusage.ErrBudget) {
		t.Fatalf("got %v", err)
	}
}

func TestStartIsNeverRefusedForAUserOnTheirOwnKey(t *testing.T) {
	creds, err := credentials.NewMemory(bytes.Repeat([]byte{1}, 32))
	if err != nil {
		t.Fatal(err)
	}
	if err := creds.Put(context.Background(), 1, "anthropic", "sk-ant-their-own-key"); err != nil {
		t.Fatal(err)
	}
	budget := &usedUp{}
	f := newFixture(t, &fakeBackend{}, nil, func(c *runner.Config) {
		c.Budget, c.Credentials, c.Backend = budget, creds, "anthropic"
	})
	if _, err := f.runner.Start(context.Background(), 1, f.session.ID, "hello"); err != nil {
		t.Fatalf("own key was refused: %v", err)
	}
	if budget.checked != 0 {
		t.Fatalf("the budget was consulted %d times for an own key", budget.checked)
	}
}

func TestCompleteReportsTheUsageOfTheCall(t *testing.T) {
	f := newFixture(t, &fakeBackend{}, nil, nil)
	text, call, err := f.runner.Complete(context.Background(), 1, "", "hi")
	if err != nil || text != "All done." {
		t.Fatalf("%q, %v", text, err)
	}
	want := nacelle.Usage{InputTokens: 10, OutputTokens: 5, Cost: 0.002}
	if call.Usage != want || call.OwnKey {
		t.Fatalf("call %+v", call)
	}
}
