package settings

import (
	"os"
	"testing"
)

func isolate(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("BULLE_TOKEN", "")
	t.Setenv("BULLE_SERVER_URL", "")
	for _, names := range vendorKeyEnv {
		for _, name := range names {
			t.Setenv(name, "")
		}
	}
}

func TestServerPrecedence(t *testing.T) {
	isolate(t)
	if err := SaveLogin("https://stored.example/", "tok"); err != nil {
		t.Fatal(err)
	}
	got, err := resolveLayers("", Config{}, Config{}, Config{})
	if err != nil || got.Server.URL != "https://stored.example" {
		t.Fatalf("stored: %q %v", got.Server.URL, err)
	}
	file := Config{Server: Server{URL: "https://file.example"}}
	got, _ = resolveLayers("", file, Config{}, Config{})
	if got.Server.URL != "https://file.example" {
		t.Fatalf("file: %q", got.Server.URL)
	}
	t.Setenv("BULLE_SERVER_URL", "https://env.example")
	got, _ = resolveLayers("", file, FromEnv(), Config{})
	if got.Server.URL != "https://env.example" {
		t.Fatalf("env: %q", got.Server.URL)
	}
}

func TestTokenEnvBeatsFileAndFileIs0600(t *testing.T) {
	isolate(t)
	if err := SaveLogin("https://s.example", "filetok"); err != nil {
		t.Fatal(err)
	}
	path, _ := tokenPath()
	if info, _ := os.Stat(path); info.Mode().Perm() != 0o600 {
		t.Fatalf("mode %v", info.Mode())
	}
	if Token() != "filetok" {
		t.Fatal("file token")
	}
	t.Setenv("BULLE_TOKEN", "envtok")
	if Token() != "envtok" {
		t.Fatal("env token")
	}
	if err := ClearLogin(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BULLE_TOKEN", "")
	if Token() != "" || StoredServerURL() != "" {
		t.Fatal("not cleared")
	}
}

func TestRouteViaServer(t *testing.T) {
	cases := []struct {
		name    string
		cfg     Config
		env     map[string]string
		wantURL string
		wantKey string
	}{
		{"routes", Config{Provider: Provider{Backend: "openai"}}, nil, "https://s.example/api/llm/openai", "tok"},
		{"api key wins", Config{Provider: Provider{Backend: "openai", APIKey: "mine"}}, nil, "", "mine"},
		{"command wins", Config{Provider: Provider{Backend: "openai", APIKeyCommand: "x"}}, nil, "", ""},
		{"base url wins", Config{Provider: Provider{Backend: "openai", BaseURL: "http://local"}}, nil, "http://local", ""},
		{"vendor env wins", Config{Provider: Provider{Backend: "anthropic"}}, map[string]string{"ANTHROPIC_API_KEY": "k"}, "", ""},
		{"other vendor env ignored", Config{Provider: Provider{Backend: "anthropic"}}, map[string]string{"OPENAI_API_KEY": "k"}, "https://s.example/api/llm/anthropic", "tok"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			isolate(t)
			t.Setenv("BULLE_TOKEN", "tok")
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			tc.cfg.Server.URL = "https://s.example/"
			RouteViaServer(&tc.cfg)
			if tc.cfg.BaseURL != tc.wantURL || tc.cfg.APIKey != tc.wantKey {
				t.Fatalf("got %q %q", tc.cfg.BaseURL, tc.cfg.APIKey)
			}
		})
	}
}

func TestRouteViaServerNeedsToken(t *testing.T) {
	isolate(t)
	cfg := Config{Provider: Provider{Backend: "openai"}, Server: Server{URL: "https://s.example"}}
	RouteViaServer(&cfg)
	if cfg.BaseURL != "" {
		t.Fatal("routed without a token")
	}
}

func TestServerKeyKnownFieldsStrict(t *testing.T) {
	path := t.TempDir() + "/c.yml"
	_ = os.WriteFile(path, []byte("server:\n  url: https://a.example\n  bogus: 1\n"), 0o600)
	if _, err := Load(path); err == nil {
		t.Fatal("unknown server field accepted")
	}
	_ = os.WriteFile(path, []byte("server:\n  url: https://a.example\n"), 0o600)
	cfg, err := Load(path)
	if err != nil || cfg.Server.URL != "https://a.example" {
		t.Fatalf("%v %q", err, cfg.Server.URL)
	}
}
