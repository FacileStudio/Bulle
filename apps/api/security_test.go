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

func TestVersionIsServed(t *testing.T) {
	rec := httptest.NewRecorder()
	testRouter().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/version", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"version"`) {
		t.Fatalf("GET /version = %d %q", rec.Code, rec.Body.String())
	}
}
