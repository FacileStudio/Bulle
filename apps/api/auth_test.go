package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/env"
	"github.com/FacileStudio/Bulle/apps/api/internal/middleware"
	"github.com/FacileStudio/Bulle/apps/api/modules/auth"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/session"
	"github.com/go-chi/chi/v5"
)

type memoryStore struct {
	mu       sync.Mutex
	sessions map[string]porte.Session
	nextID   int64
}

func newMemoryStore() *memoryStore {
	return &memoryStore{sessions: map[string]porte.Session{}}
}

func (m *memoryStore) Create(_ context.Context, created porte.Session) (porte.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextID++
	created.ID = m.nextID
	m.sessions[created.TokenHash] = created
	return created, nil
}

func (m *memoryStore) Find(_ context.Context, tokenHash string) (porte.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	found, ok := m.sessions[tokenHash]
	if !ok {
		return porte.Session{}, porte.ErrNotFound
	}
	return found, nil
}

func (m *memoryStore) Touch(_ context.Context, tokenHash string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	touched := m.sessions[tokenHash]
	touched.LastUsedAt = at
	m.sessions[tokenHash] = touched
	return nil
}

func (m *memoryStore) Delete(_ context.Context, tokenHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, tokenHash)
	return nil
}

func (m *memoryStore) DeleteByUser(context.Context, int64) (int64, error)        { return 0, nil }
func (m *memoryStore) DeleteLogins(context.Context, int64, int64) (int64, error) { return 0, nil }
func (m *memoryStore) ListByUser(context.Context, int64) ([]porte.Session, error) {
	return nil, nil
}
func (m *memoryStore) DeleteByID(context.Context, int64, int64) error          { return porte.ErrNotFound }
func (m *memoryStore) DeleteExpired(context.Context, time.Time) (int64, error) { return 0, nil }

type fixedResolver struct{}

func (fixedResolver) IdentityForUser(_ context.Context, userID int64) (authcontext.Identity, error) {
	return authcontext.Identity{UserID: userID, Email: "someone@facile.studio"}, nil
}

func protectedRouter(t *testing.T) (chi.Router, string) {
	t.Helper()
	manager, err := session.New(porte.Config{SuccessURL: "http://bulle.test/"}, session.Deps{
		Sessions: newMemoryStore(),
		Logger:   slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatalf("session.New: %v", err)
	}
	token, _, err := manager.Issue(context.Background(), 7, "")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	router := chi.NewRouter()
	guard := middleware.RequireAuth(manager, fixedResolver{})
	router.With(guard).Get("/read", ok)
	router.With(guard).Post("/write", ok)
	return router, token
}

func fakeIssuer() string {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"issuer":                 "http://" + r.Host,
			"authorization_endpoint": "http://" + r.Host + "/auth",
			"token_endpoint":         "http://" + r.Host + "/token",
			"jwks_uri":               "http://" + r.Host + "/keys",
		})
	}))
	return server.URL
}

func send(router chi.Router, method, path string, headers map[string]string) int {
	request := httptest.NewRequest(method, path, nil)
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder.Code
}

func TestCookieMutationIsRefusedWithoutTheCSRFHeader(t *testing.T) {
	router, token := protectedRouter(t)
	cookie := map[string]string{"Cookie": porte.SessionCookieName + "=" + token}

	if code := send(router, http.MethodGet, "/read", cookie); code != http.StatusOK {
		t.Fatalf("cookie read = %d, want 200", code)
	}
	if code := send(router, http.MethodPost, "/write", cookie); code != http.StatusForbidden {
		t.Fatalf("cookie write without CSRF = %d, want 403", code)
	}
	withCSRF := map[string]string{
		"Cookie":             porte.SessionCookieName + "=" + token,
		porte.CSRFHeaderName: "1",
	}
	if code := send(router, http.MethodPost, "/write", withCSRF); code != http.StatusOK {
		t.Fatalf("cookie write with CSRF = %d, want 200", code)
	}
}

func TestCookieIsReadBeforeTheAuthorizationHeader(t *testing.T) {
	router, token := protectedRouter(t)
	headers := map[string]string{
		"Cookie":        porte.SessionCookieName + "=" + token,
		"Authorization": "Bearer not-a-real-token",
	}
	if code := send(router, http.MethodGet, "/read", headers); code != http.StatusOK {
		t.Fatalf("valid cookie with a junk bearer = %d, want 200", code)
	}
}

func TestBearerMutationNeedsNoCSRFHeader(t *testing.T) {
	router, token := protectedRouter(t)
	headers := map[string]string{"Authorization": "Bearer " + token}
	if code := send(router, http.MethodPost, "/write", headers); code != http.StatusOK {
		t.Fatalf("bearer write = %d, want 200", code)
	}
}

func TestUnverifiedEmailIsRefusedBeforeAnyAccountIsTouched(t *testing.T) {
	users := auth.NewUserStore(nil, nil)
	_, err := users.UpsertFromOIDC(context.Background(), porte.Claims{
		Provider:      "https://sso.facile.studio",
		Subject:       "abc",
		Email:         "someone@facile.studio",
		EmailVerified: false,
	})
	if err == nil {
		t.Fatal("an unverified email was accepted")
	}
}

func TestPasswordRoutesDoNotExist(t *testing.T) {
	issuer := fakeIssuer()
	router, err := buildRouter(routerDeps{
		Env: env.Config{Porte: porte.Config{
			Issuer:       issuer,
			ClientID:     "bulle",
			ClientSecret: "secret",
			RedirectURL:  "https://bulle.facile.studio/api/auth/oidc/callback",
			SuccessURL:   "https://bulle.facile.studio/",
		}},
		Logger: slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	for _, path := range []string{"/api/auth/login", "/api/auth/register"} {
		code := send(router, http.MethodPost, path, nil)
		if code != http.StatusNotFound && code != http.StatusMethodNotAllowed {
			t.Fatalf("%s = %d, want 404 or 405", path, code)
		}
	}
}
