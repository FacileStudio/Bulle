package suite

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/porte"
)

type memIdentities struct {
	mu    sync.Mutex
	rows  map[int64]porte.StoredIdentity
	saves int
}

func (m *memIdentities) ListByUser(_ context.Context, userID int64) ([]porte.StoredIdentity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if row, ok := m.rows[userID]; ok {
		return []porte.StoredIdentity{row}, nil
	}
	return nil, nil
}

func (m *memIdentities) Save(_ context.Context, identity porte.StoredIdentity) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rows[identity.UserID] = identity
	m.saves++
	return nil
}

type world struct {
	registre *httptest.Server
	app      *httptest.Server
	ids      *memIdentities
	cfg      Config

	mu        sync.Mutex
	exchanges int
	refreshes int
	rejected  map[string]bool
	revoked   map[string]bool
	unknown   map[string]bool
	readOnly  bool
	appCalls  int
	writes    []recorded
}

func newWorld(t *testing.T) *world {
	t.Helper()
	w := &world{
		ids:      &memIdentities{rows: map[int64]porte.StoredIdentity{}},
		rejected: map[string]bool{}, revoked: map[string]bool{}, unknown: map[string]bool{},
	}
	w.registre = httptest.NewServer(http.HandlerFunc(w.serveRegistre))
	w.app = httptest.NewServer(http.HandlerFunc(w.serveApp))
	t.Cleanup(w.registre.Close)
	t.Cleanup(w.app.Close)
	w.cfg = Config{Issuer: w.registre.URL, ClientID: "bulle", ClientSecret: "s3cret", Apps: map[string]AppConfig{}}
	w.cfg.Apps[AppSablier] = AppConfig{URL: w.app.URL, Audience: "aud-" + AppSablier}
	return w
}

func (w *world) addUser(id int64, expiry time.Time) {
	w.ids.rows[id] = porte.StoredIdentity{
		UserID: id, Provider: w.cfg.Issuer, Subject: fmt.Sprint("sub-", id),
		Tokens: porte.TokenSet{AccessToken: fmt.Sprintf("at:%d:0", id), RefreshToken: fmt.Sprintf("rt:%d:0", id), Expiry: expiry},
	}
}

func (w *world) tokens() *Tokens { return NewTokens(w.cfg, w.ids, nil) }

func (w *world) count(field *int) int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return *field
}

func (w *world) serveRegistre(rw http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/.well-known/openid-configuration" {
		_ = json.NewEncoder(rw).Encode(map[string]string{"token_endpoint": w.registre.URL + "/token"})
		return
	}
	id, secret, ok := r.BasicAuth()
	if !ok || id != "bulle" || secret != "s3cret" {
		writeOAuthError(rw, http.StatusUnauthorized, "invalid_client")
		return
	}
	_ = r.ParseForm()
	w.mu.Lock()
	defer w.mu.Unlock()
	if r.Form.Get("grant_type") == "urn:ietf:params:oauth:grant-type:token-exchange" {
		w.exchange(rw, r)
		return
	}
	w.refresh(rw, r.Form.Get("refresh_token"))
}

func (w *world) refresh(rw http.ResponseWriter, refreshToken string) {
	if w.revoked[refreshToken] {
		writeOAuthError(rw, http.StatusBadRequest, "invalid_grant")
		return
	}
	w.refreshes++
	user, gen := splitToken(refreshToken)
	_ = json.NewEncoder(rw).Encode(map[string]any{
		"access_token": fmt.Sprintf("at:%s:%d", user, gen+1), "refresh_token": fmt.Sprintf("rt:%s:%d", user, gen+1), "expires_in": 3600,
	})
}

func (w *world) exchange(rw http.ResponseWriter, r *http.Request) {
	subject := r.Form.Get("subject_token")
	if w.rejected[subject] || r.Form.Get("subject_token_type") != "urn:ietf:params:oauth:token-type:access_token" {
		writeOAuthError(rw, http.StatusBadRequest, "invalid_grant")
		return
	}
	if w.readOnly && strings.Contains(r.Form.Get("scope"), ".write") {
		writeOAuthError(rw, http.StatusBadRequest, "invalid_scope")
		return
	}
	w.exchanges++
	user, _ := splitToken(subject)
	_ = json.NewEncoder(rw).Encode(map[string]any{
		"access_token": fmt.Sprintf("jwt:%s:%s:%s:%d", user, r.Form.Get("audience"), strings.ReplaceAll(r.Form.Get("scope"), " ", "+"), w.exchanges),
		"expires_in":   300,
	})
}

func splitToken(token string) (string, int) {
	parts := strings.Split(token, ":")
	var gen int
	_, _ = fmt.Sscan(parts[2], &gen)
	return parts[1], gen
}

func writeOAuthError(rw http.ResponseWriter, status int, code string) {
	rw.WriteHeader(status)
	_ = json.NewEncoder(rw).Encode(map[string]string{"error": code})
}
