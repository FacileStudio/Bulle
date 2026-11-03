package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

func exerciseSearchQuota(t *testing.T, s store.Store, owner, stranger int64) {
	t.Helper()
	ctx := context.Background()
	day := time.Date(2026, 11, 3, 23, 0, 0, 0, time.UTC)
	for i, want := range []bool{true, true, false} {
		if got, err := s.TakeSearch(ctx, owner, day, 2); err != nil || got != want {
			t.Fatalf("take %d: %v %v, want %v", i, got, err, want)
		}
	}
	if got, _ := s.TakeSearch(ctx, stranger, day, 2); !got {
		t.Fatal("one user's searches counted against another")
	}
	if got, _ := s.TakeSearch(ctx, owner, day.Add(2*time.Hour), 2); !got {
		t.Fatal("the next day did not start a fresh count")
	}
	if got, _ := s.TakeSearch(ctx, owner, day.AddDate(0, 0, 2), 0); got {
		t.Fatal("a zero limit allowed a search")
	}
}
