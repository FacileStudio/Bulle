package ai_test

import (
	"crypto/sha256"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/middleware"
	"github.com/FacileStudio/Bulle/apps/api/modules/ai"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/session"
	"github.com/go-chi/chi/v5"
)

func cookieSession(next, bearer http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := r.Cookie(porte.SessionCookieName); err != nil {
			bearer.ServeHTTP(w, r)
			return
		}
		ctx := porte.WithIdentity(r.Context(), porte.Identity{UserID: 7, Subject: "sub-7"})
		next.ServeHTTP(w, r.WithContext(authcontext.WithIdentity(ctx, authcontext.Identity{UserID: 7, Email: member})))
	})
}

func exchangedServer(t *testing.T, decider ai.Decider, writer ai.Summarizer) *httptest.Server {
	t.Helper()
	manager, err := session.New(porte.Config{SuccessURL: "https://bulle.test/", ScopePrefix: "bulle"}, session.Deps{Sessions: noSessions{}})
	if err != nil {
		t.Fatal(err)
	}
	manager.WithJWT(jwts{
		toileWrite: {UserID: 7, Subject: "sub-7", Scopes: []string{"bulle.read", "bulle.write"}, Actor: "sablier"},
		userToken:  {UserID: 7, Subject: "sub-7", Scopes: []string{"bulle.read", "bulle.write"}},
	})
	auth := middleware.RequireAuth(manager, byID{})
	sum := sha256.Sum256([]byte(serviceKey))
	router := chi.NewRouter()
	ai.RegisterRoutes(router, ai.Config{
		Decider:    decider,
		Summarizer: writer,
		RequireAuth: func(next http.Handler) http.Handler {
			return cookieSession(next, auth(next))
		},
		Callers: map[string][]byte{"sablier": sum[:]},
		Users:   users{},
		Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	return srv
}

func call(t *testing.T, srv *httptest.Server, path string, headers map[string]string, body string) int {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, srv.URL+path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	return res.StatusCode
}

func TestDecideTakesATokenAnAppExchanged(t *testing.T) {
	var calls atomic.Int32
	srv := exchangedServer(t, stubJev(t, &calls), nil)
	status := call(t, srv, "/ai/decide", map[string]string{"Authorization": "Bearer " + toileWrite}, question)
	if status != http.StatusOK || calls.Load() != 1 {
		t.Fatalf("status %d, model calls %d", status, calls.Load())
	}
}

func TestSummarizeTakesATokenAnAppExchanged(t *testing.T) {
	writer := &fakeWriter{answer: `{"summary":"Go 1.18 added generics [1]."}`}
	srv := exchangedServer(t, nil, writer)
	status := call(t, srv, "/ai/summarize", map[string]string{"Authorization": "Bearer " + toileWrite}, twoSources)
	if status != http.StatusOK || writer.calls() != 1 {
		t.Fatalf("status %d, model calls %d", status, writer.calls())
	}
}

func TestAIRoutesRefuseASessionAndATokenWithoutAnActor(t *testing.T) {
	var calls atomic.Int32
	writer := &fakeWriter{answer: `{"summary":"x"}`}
	srv := exchangedServer(t, stubJev(t, &calls), writer)
	for path, body := range map[string]string{"/ai/decide": question, "/ai/summarize": twoSources} {
		for name, headers := range map[string]map[string]string{
			"cookie session": {"Cookie": "session=abc"},
			"no actor":       {"Authorization": "Bearer " + userToken},
		} {
			if status := call(t, srv, path, headers, body); status != http.StatusForbidden {
				t.Errorf("%s %s: status %d, want 403", path, name, status)
			}
		}
		if status := call(t, srv, path, nil, body); status != http.StatusUnauthorized {
			t.Errorf("%s signed out: status %d, want 401", path, status)
		}
	}
	if calls.Load() != 0 || writer.calls() != 0 {
		t.Fatal("a refused call reached a model")
	}
}

func TestCallerKeyStillWorksBesideRequireAuth(t *testing.T) {
	var calls atomic.Int32
	writer := &fakeWriter{answer: `{"summary":"Go 1.18 added generics [1]."}`}
	srv := exchangedServer(t, stubJev(t, &calls), writer)
	headers := map[string]string{"Authorization": "Bearer " + serviceKey, "X-Bulle-Acting-For": member}
	if status := call(t, srv, "/ai/decide", headers, question); status != http.StatusOK {
		t.Fatalf("decide: status %d", status)
	}
	if status := call(t, srv, "/ai/summarize", headers, twoSources); status != http.StatusOK {
		t.Fatalf("summarize: status %d", status)
	}
}
