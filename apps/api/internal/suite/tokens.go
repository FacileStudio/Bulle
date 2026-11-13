package suite

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/exchange"
)

type Identities interface {
	ListByUser(ctx context.Context, userID int64) ([]porte.StoredIdentity, error)
	Save(ctx context.Context, identity porte.StoredIdentity) error
}

// Tokens mints the delegated tokens Bulle calls the suite with, through
// porte/exchange: one Registre refresh per user at a time across every app,
// the identity that still holds a refresh token, and an in-memory cache.
type Tokens struct {
	cfg    Config
	now    func() time.Time
	ex     *exchange.Exchanger
	err    error
	outage outage
}

func NewTokens(cfg Config, ids Identities, client *http.Client) *Tokens {
	t := &Tokens{cfg: cfg, now: time.Now}
	t.ex, t.err = exchange.New(exchange.Config{Issuer: cfg.Issuer, ClientID: cfg.ClientID, ClientSecret: cfg.ClientSecret,
		Identities: ids, Client: client, Now: func() time.Time { return t.now() }})
	return t
}

// ForMCP is the user's token for app's MCP server: the resource URI as the
// audience, which Registre resolves to the app and binds the token to. It
// asks for write and falls back to read when Registre refuses write. A server
// with no write scope, such as sonar's, is only ever asked for read.
func (t *Tokens) ForMCP(ctx context.Context, userID int64, app string) (string, error) {
	resource := t.cfg.MCP[app].Resource
	if !t.cfg.MCP[app].Enabled() {
		return "", fmt.Errorf("suite: %s is not configured", app)
	}
	if readOnlyMCP[app] {
		return t.mint(ctx, userID, readTarget(resource, app))
	}
	token, err := t.mint(ctx, userID, writeTarget(resource, app))
	var refused *exchange.RefusedError
	if errors.As(err, &refused) && refused.Code == "invalid_scope" {
		return t.mint(ctx, userID, readTarget(resource, app))
	}
	return token, err
}

// mint exchanges a token for target, and rides out a Registre that does not
// answer: a redeploy leaves Traefik answering 404 for about 30 s.
func (t *Tokens) mint(ctx context.Context, userID int64, target exchange.Target) (string, error) {
	if t.err != nil {
		return "", t.err
	}
	try := func() (string, error) { return t.exchange(ctx, userID, target) }
	token, err := try()
	if !transientExchange(err) {
		t.outage.end()
		return token, err
	}
	return t.outage.ride(ctx, t.now, try, err)
}

func (t *Tokens) exchange(ctx context.Context, userID int64, target exchange.Target) (string, error) {
	token, err := t.ex.Token(ctx, userID, target)
	if errors.Is(err, exchange.ErrReconnect) {
		return "", ErrNotConnected
	}
	return token, err
}

// InvalidateMCP drops one user's cached tokens for app's MCP server.
func (t *Tokens) InvalidateMCP(userID int64, app string) {
	t.forget(userID, t.cfg.MCP[app].Resource, app)
}

func (t *Tokens) forget(userID int64, audience, app string) {
	if t.err != nil {
		return
	}
	t.ex.Forget(userID, readTarget(audience, app))
	t.ex.Forget(userID, writeTarget(audience, app))
}

// Forget drops everything cached for a user, on sign-out and erasure.
func (t *Tokens) Forget(userID int64) {
	if t.err == nil {
		t.ex.ForgetUser(userID)
	}
}

func readTarget(audience, app string) exchange.Target {
	return exchange.Target{Audience: audience, Scopes: []string{Scope(app)}}
}

func writeTarget(audience, app string) exchange.Target {
	return exchange.Target{Audience: audience, Scopes: []string{Scope(app), app + ".write"}}
}
