package suite

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/FacileStudio/porte"
)

const cacheMargin = 30 * time.Second

type Identities interface {
	ListByUser(ctx context.Context, userID int64) ([]porte.StoredIdentity, error)
	Save(ctx context.Context, identity porte.StoredIdentity) error
}

type cacheKey struct {
	user  int64
	app   string
	write bool
	mcp   bool
}

type Tokens struct {
	cfg     Config
	ids     Identities
	reg     *Registre
	now     func() time.Time
	flights flightGroup

	mu    sync.Mutex
	cache map[cacheKey]Delegated
}

func NewTokens(cfg Config, ids Identities, client *http.Client) *Tokens {
	t := &Tokens{cfg: cfg, ids: ids, now: time.Now, cache: map[cacheKey]Delegated{}}
	t.reg = newRegistre(cfg, client, func() time.Time { return t.now() })
	return t
}

func (t *Tokens) For(ctx context.Context, userID int64, app string) (string, error) {
	return t.token(ctx, cacheKey{user: userID, app: app})
}

func (t *Tokens) ForWrite(ctx context.Context, userID int64, app string) (string, error) {
	return t.token(ctx, cacheKey{user: userID, app: app, write: true})
}

func (t *Tokens) token(ctx context.Context, key cacheKey) (string, error) {
	if !t.configured(key) {
		return "", fmt.Errorf("suite: %s is not configured", key.app)
	}
	if token, ok := t.cached(key); ok {
		return token, nil
	}
	flight := fmt.Sprintf("mint/%d/%s/%t/%t", key.user, key.app, key.write, key.mcp)
	value, err := t.flights.do(flight, func() (string, error) {
		return t.mint(context.WithoutCancel(ctx), key)
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

func (t *Tokens) Invalidate(userID int64, app string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.cache, cacheKey{user: userID, app: app})
	delete(t.cache, cacheKey{user: userID, app: app, write: true})
}

func (t *Tokens) Forget(userID int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for key := range t.cache {
		if key.user == userID {
			delete(t.cache, key)
		}
	}
}

func (t *Tokens) cached(key cacheKey) (string, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	entry, ok := t.cache[key]
	if !ok || !t.now().Add(cacheMargin).Before(entry.Expiry) {
		return "", false
	}
	return entry.Token, true
}

func (t *Tokens) mint(ctx context.Context, key cacheKey) (string, error) {
	if token, ok := t.cached(key); ok {
		return token, nil
	}
	delegated, err := t.exchange(ctx, key, false)
	if errors.Is(err, ErrSubjectRejected) {
		delegated, err = t.exchange(ctx, key, true)
	}
	if errors.Is(err, ErrSubjectRejected) {
		return "", ErrNotConnected
	}
	if err != nil {
		return "", err
	}
	t.mu.Lock()
	t.cache[key] = delegated
	t.mu.Unlock()
	return delegated.Token, nil
}

func (t *Tokens) exchange(ctx context.Context, key cacheKey, forceRefresh bool) (Delegated, error) {
	subject, err := t.subject(ctx, key.user, forceRefresh)
	if err != nil {
		return Delegated{}, err
	}
	if key.mcp {
		return t.reg.ExchangeResource(ctx, subject, key.app, t.cfg.MCP[key.app].Resource)
	}
	return t.reg.Exchange(ctx, subject, key.app, key.write)
}
