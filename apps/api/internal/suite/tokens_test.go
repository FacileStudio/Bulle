package suite

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/porte"
)

var future = time.Now().Add(time.Hour)

func TestTokensCacheHit(t *testing.T) {
	w := sablierWorld(t)
	w.addUser(1, future)
	tokens := w.tokens()
	first, err := tokens.ForMCP(context.Background(), 1, AppSablier)
	if err != nil {
		t.Fatal(err)
	}
	second, _ := tokens.ForMCP(context.Background(), 1, AppSablier)
	if first != second || w.count(&w.exchanges) != 1 {
		t.Fatalf("want one exchange and a shared token, got %d, %q vs %q", w.count(&w.exchanges), first, second)
	}
	if !strings.Contains(first, ":aud-sablier:sablier.read+sablier.write:") {
		t.Fatalf("token is not scoped to sablier read: %q", first)
	}
}

func TestTokensCacheExpiry(t *testing.T) {
	w := sablierWorld(t)
	w.addUser(1, future)
	tokens := w.tokens()
	_, _ = tokens.ForMCP(context.Background(), 1, AppSablier)
	tokens.now = func() time.Time { return time.Now().Add(6 * time.Minute) }
	_, _ = tokens.ForMCP(context.Background(), 1, AppSablier)
	if got := w.count(&w.exchanges); got != 2 {
		t.Fatalf("expired entry must be re-minted, exchanges = %d", got)
	}
}

func TestTokensSingleflight(t *testing.T) {
	w := sablierWorld(t)
	w.addUser(1, future)
	tokens := w.tokens()
	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			if _, err := tokens.ForMCP(context.Background(), 1, AppSablier); err != nil {
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
	w := sablierWorld(t)
	w.addUser(1, future)
	tokens := w.tokens()
	w.addUser(2, future)
	_, _ = tokens.ForMCP(context.Background(), 1, AppSablier)
	_, _ = tokens.ForMCP(context.Background(), 2, AppSablier)
	tokens.InvalidateMCP(1, AppSablier)
	_, _ = tokens.ForMCP(context.Background(), 1, AppSablier)
	_, _ = tokens.ForMCP(context.Background(), 2, AppSablier)
	if got := w.count(&w.exchanges); got != 3 {
		t.Fatalf("invalidate must drop only one entry, exchanges = %d", got)
	}
	tokens.Forget(1)
	_, _ = tokens.ForMCP(context.Background(), 1, AppSablier)
	if got := w.count(&w.exchanges); got != 4 {
		t.Fatalf("forget must drop the user's entries, exchanges = %d", got)
	}
}

func TestTokensRefreshWritesRotationBack(t *testing.T) {
	w := sablierWorld(t)
	w.addUser(1, time.Now().Add(-time.Minute))
	if _, err := w.tokens().ForMCP(context.Background(), 1, AppSablier); err != nil {
		t.Fatal(err)
	}
	saved := w.ids.rows[1].Tokens
	if saved.AccessToken != "at:1:1" || saved.RefreshToken != "rt:1:1" || w.ids.saves != 1 {
		t.Fatalf("rotated pair not stored: %+v, saves %d", saved, w.ids.saves)
	}
}

func TestTokensConcurrentAppsRefreshOnce(t *testing.T) {
	w := sablierWorld(t)
	w.addUser(1, time.Now().Add(-time.Minute))
	apps := []string{AppAgenda, AppCourrier, AppNuage, AppOpus}
	w.cfg.MCP = map[string]MCPConfig{}
	for _, app := range apps {
		w.cfg.MCP[app] = MCPConfig{URL: w.app.URL + "/" + app, Resource: w.app.URL + "/" + app}
	}
	tokens := w.tokens()
	var wg sync.WaitGroup
	for _, app := range apps {
		wg.Go(func() {
			if _, err := tokens.ForMCP(context.Background(), 1, app); err != nil {
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
	w := sablierWorld(t)
	w.addUser(1, future)
	w.rejected["at:1:0"] = true
	token, err := w.tokens().ForMCP(context.Background(), 1, AppSablier)
	if err != nil || token == "" {
		t.Fatalf("a revoked-but-unexpired token must trigger one refresh, got %v", err)
	}
	if w.count(&w.refreshes) != 1 {
		t.Fatal("expected a forced refresh")
	}
}

func TestTokensErrors(t *testing.T) {
	w := sablierWorld(t)
	w.addUser(1, time.Now().Add(-time.Minute))
	w.revoked["rt:1:0"] = true
	w.addUser(2, future)
	w.rejected["at:2:0"] = true
	w.revoked["rt:2:0"] = true
	tokens := w.tokens()
	cases := map[string]int64{"no identity": 9, "refresh rejected": 1, "subject rejected twice": 2}
	for name, user := range cases {
		if _, err := tokens.ForMCP(context.Background(), user, AppSablier); !errors.Is(err, ErrNotConnected) {
			t.Errorf("%s: want ErrNotConnected, got %v", name, err)
		}
	}
	if _, err := tokens.ForMCP(context.Background(), 1, "mail"); err == nil {
		t.Error("an unconfigured app must be refused")
	}
}

type leftoverIdentities struct {
	*memIdentities
	issuer string
}

func (l leftoverIdentities) ListByUser(ctx context.Context, userID int64) ([]porte.StoredIdentity, error) {
	live, err := l.memIdentities.ListByUser(ctx, userID)
	stale := porte.StoredIdentity{UserID: userID, Provider: l.issuer, Subject: "leftover",
		Tokens: porte.TokenSet{AccessToken: "dead", Expiry: time.Now().Add(-time.Hour)}}
	return append([]porte.StoredIdentity{stale}, live...), err
}

func TestTokensSkipALeftoverIdentityWithoutARefreshToken(t *testing.T) {
	w := sablierWorld(t)
	w.addUser(1, time.Now().Add(-time.Minute))
	tokens := NewTokens(w.cfg, leftoverIdentities{memIdentities: w.ids, issuer: w.cfg.Issuer}, nil)
	token, err := tokens.ForMCP(context.Background(), 1, AppSablier)
	if err != nil || token == "" {
		t.Fatalf("the identity that holds a refresh token must be used, got %v", err)
	}
	if w.count(&w.refreshes) != 1 || w.ids.rows[1].Tokens.RefreshToken != "rt:1:1" {
		t.Fatalf("the live grant must be refreshed and its rotation stored: %+v", w.ids.rows[1].Tokens)
	}
}
