package suite

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

var future = time.Now().Add(time.Hour)

func TestTokensCacheHit(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	tokens := w.tokens()
	first, err := tokens.For(context.Background(), 1, AppAgenda)
	if err != nil {
		t.Fatal(err)
	}
	second, _ := tokens.For(context.Background(), 1, AppAgenda)
	if first != second || w.count(&w.exchanges) != 1 {
		t.Fatalf("want one exchange and a shared token, got %d, %q vs %q", w.count(&w.exchanges), first, second)
	}
	if !strings.Contains(first, ":aud-agenda:agenda.read:") {
		t.Fatalf("token is not scoped to agenda read: %q", first)
	}
}

func TestTokensCacheExpiry(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	tokens := w.tokens()
	_, _ = tokens.For(context.Background(), 1, AppAgenda)
	tokens.now = func() time.Time { return time.Now().Add(6 * time.Minute) }
	_, _ = tokens.For(context.Background(), 1, AppAgenda)
	if got := w.count(&w.exchanges); got != 2 {
		t.Fatalf("expired entry must be re-minted, exchanges = %d", got)
	}
}

func TestTokensSingleflight(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	tokens := w.tokens()
	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			if _, err := tokens.For(context.Background(), 1, AppOpus); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if got := w.count(&w.exchanges); got != 1 {
		t.Fatalf("50 concurrent misses must share one exchange, got %d", got)
	}
}

func TestTokensForgetAndInvalidate(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	tokens := w.tokens()
	_, _ = tokens.For(context.Background(), 1, AppOpus)
	_, _ = tokens.For(context.Background(), 1, AppSablier)
	tokens.Invalidate(1, AppOpus)
	_, _ = tokens.For(context.Background(), 1, AppOpus)
	_, _ = tokens.For(context.Background(), 1, AppSablier)
	if got := w.count(&w.exchanges); got != 3 {
		t.Fatalf("invalidate must drop only one entry, exchanges = %d", got)
	}
	tokens.Forget(1)
	_, _ = tokens.For(context.Background(), 1, AppSablier)
	if got := w.count(&w.exchanges); got != 4 {
		t.Fatalf("forget must drop the user's entries, exchanges = %d", got)
	}
}

func TestTokensRefreshWritesRotationBack(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, time.Now().Add(-time.Minute))
	if _, err := w.tokens().For(context.Background(), 1, AppAgenda); err != nil {
		t.Fatal(err)
	}
	saved := w.ids.rows[1].Tokens
	if saved.AccessToken != "at:1:1" || saved.RefreshToken != "rt:1:1" || w.ids.saves != 1 {
		t.Fatalf("rotated pair not stored: %+v, saves %d", saved, w.ids.saves)
	}
}

func TestTokensConcurrentAppsRefreshOnce(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, time.Now().Add(-time.Minute))
	tokens := w.tokens()
	var wg sync.WaitGroup
	for _, app := range appNames {
		wg.Go(func() {
			if _, err := tokens.For(context.Background(), 1, app); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if got := w.count(&w.refreshes); got != 1 {
		t.Fatalf("one rotation per user, got %d refreshes", got)
	}
}

func TestTokensSubjectRejectedForcesRefresh(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	w.rejected["at:1:0"] = true
	token, err := w.tokens().For(context.Background(), 1, AppAgenda)
	if err != nil || token == "" {
		t.Fatalf("a revoked-but-unexpired token must trigger one refresh, got %v", err)
	}
	if w.count(&w.refreshes) != 1 {
		t.Fatal("expected a forced refresh")
	}
}

func TestTokensErrors(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, time.Now().Add(-time.Minute))
	w.revoked["rt:1:0"] = true
	w.addUser(2, future)
	w.rejected["at:2:0"] = true
	w.revoked["rt:2:0"] = true
	tokens := w.tokens()
	cases := map[string]int64{"no identity": 9, "refresh rejected": 1, "subject rejected twice": 2}
	for name, user := range cases {
		if _, err := tokens.For(context.Background(), user, AppAgenda); !errors.Is(err, ErrNotConnected) {
			t.Errorf("%s: want ErrNotConnected, got %v", name, err)
		}
	}
	if _, err := tokens.For(context.Background(), 1, "mail"); err == nil {
		t.Error("an unconfigured app must be refused")
	}
}
