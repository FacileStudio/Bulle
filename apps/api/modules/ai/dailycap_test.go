package ai

import (
	"testing"
	"time"
)

func TestDailyCapResetsEachUTCDay(t *testing.T) {
	now := time.Date(2026, 11, 11, 23, 0, 0, 0, time.UTC)
	c := &dailyCap{limit: 2, now: func() time.Time { return now }}
	for i, want := range []bool{true, true, false} {
		if got := c.take(1); got != want {
			t.Fatalf("completion %d: allowed %v, want %v", i+1, got, want)
		}
	}
	if !c.take(2) {
		t.Fatal("one user's cap blocked another")
	}
	now = now.Add(2 * time.Hour)
	if !c.take(1) {
		t.Fatal("the cap did not reset the next day")
	}
}
