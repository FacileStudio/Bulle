package settings

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRemoteConfigIsOneLayerAndMayNotMoveTheServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/cli/config" {
			t.Errorf("path = %q, want /api/cli/config", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer tok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"config":"limits:\n  max_concurrency: 4\nserver:\n  url: https://elsewhere.example\n"}`))
	}))
	t.Cleanup(srv.Close)

	cfg := RemoteConfig(srv.URL, "tok")
	if cfg.MaxConcurrency == nil || *cfg.MaxConcurrency != 4 {
		t.Errorf("max concurrency = %v, want the saved 4", cfg.MaxConcurrency)
	}
	if cfg.Server.URL != "" {
		t.Errorf("server url = %q, want it dropped: the token goes to the address the machine already holds", cfg.Server.URL)
	}

	if got := RemoteConfig(srv.URL, ""); got.MaxConcurrency != nil {
		t.Errorf("no token still produced a layer: %+v", got)
	}
	if got := RemoteConfig("http://127.0.0.1:1", "tok"); got.MaxConcurrency != nil {
		t.Errorf("an unreachable server still produced a layer: %+v", got)
	}
}
