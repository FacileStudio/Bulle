package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/env"
)

func TestSecurityHeadersCarryTheClientScriptHashes(t *testing.T) {
	dir := t.TempDir()
	page := `<!doctype html><html><body><script>window.boot = 1;</script></body></html>`
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(page), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLIENT_DIR", dir)

	router, err := buildRouter(routerDeps{Env: env.Config{}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("GET / = %d, want 200", rec.Code)
	}
	policy := rec.Header().Get("Content-Security-Policy-Report-Only")
	if !strings.Contains(policy, "'sha256-") {
		t.Errorf("report-only policy lacks the inline script hash: %q", policy)
	}
	if !strings.Contains(rec.Header().Get("Content-Security-Policy"), "frame-ancestors 'none'") {
		t.Errorf("anti-framing is not enforced: %q", rec.Header().Get("Content-Security-Policy"))
	}
}

func clientHeaders(t *testing.T) func(path string) http.Header {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<!doctype html><html><body></body></html>`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLIENT_DIR", dir)
	router, err := buildRouter(routerDeps{Env: env.Config{}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	if err != nil {
		t.Fatal(err)
	}
	return func(path string) http.Header {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		return rec.Header()
	}
}

func TestEmbedRoutesMayBeFramedByTheSuite(t *testing.T) {
	embed := clientHeaders(t)("/embed/chat?app=docs&id=1")
	want := "frame-ancestors https://docs.facile.studio https://toile.facile.studio https://feuilles.facile.studio"
	if got := embed.Get("Content-Security-Policy"); !strings.Contains(got, want) || strings.Contains(got, "frame-ancestors 'none'") {
		t.Errorf("/embed/chat CSP = %q, want %q", got, want)
	}
	if got := embed.Get("X-Frame-Options"); got != "" {
		t.Errorf("/embed/chat still sends X-Frame-Options %q", got)
	}
}

func TestEveryOtherRouteRefusesFraming(t *testing.T) {
	serve := clientHeaders(t)
	for _, path := range []string{"/", "/s/abc", "/embedded", "/version"} {
		header := serve(path)
		if got := header.Get("Content-Security-Policy"); !strings.Contains(got, "frame-ancestors 'none'") {
			t.Errorf("%s CSP = %q, want frame-ancestors 'none'", path, got)
		}
		if got := header.Get("X-Frame-Options"); got != "DENY" {
			t.Errorf("%s X-Frame-Options = %q, want DENY", path, got)
		}
	}
}

func TestVersionIsServed(t *testing.T) {
	rec := httptest.NewRecorder()
	testRouter().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/version", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"version"`) {
		t.Fatalf("GET /version = %d %q", rec.Code, rec.Body.String())
	}
}
