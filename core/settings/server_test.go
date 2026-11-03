package settings

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

func isolate(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("BULLE_TOKEN", "")
	t.Setenv("BULLE_SERVER_URL", "")
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

func llmServer(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/llm" || r.Header.Get("Authorization") != "Bearer tok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"backend":"openrouter","model":"a/default","models":["a/default","b/other"],"persona":"You are Nook."}`))
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestUseServerTakesTheProviderFromTheServer(t *testing.T) {
	isolate(t)
	t.Setenv("BULLE_TOKEN", "tok")
	url := llmServer(t)
	cfg := Config{Server: Server{URL: url + "/"}}
	if err := UseServer(&cfg); err != nil {
		t.Fatal(err)
	}
	want := Provider{Backend: "openrouter", Model: "a/default", BaseURL: url + "/api/llm/openrouter", APIKey: "tok", Models: []string{"a/default", "b/other"}, Persona: "You are Nook."}
	if !reflect.DeepEqual(cfg.Provider, want) {
		t.Fatalf("got %+v, want %+v", cfg.Provider, want)
	}
	cfg = Config{Server: Server{URL: url}, Provider: Provider{Model: "b/other"}}
	if err := UseServer(&cfg); err != nil || cfg.Model != "b/other" {
		t.Fatalf("offered model: %q %v", cfg.Model, err)
	}
	cfg = Config{Server: Server{URL: url}, Provider: Provider{Model: "c/unknown"}}
	if err := UseServer(&cfg); err == nil || !strings.Contains(err.Error(), "not offered") {
		t.Fatalf("unknown model err = %v", err)
	}
}

func TestUseServerRefusesWithoutALogin(t *testing.T) {
	isolate(t)
	cfg := Config{Server: Server{URL: "https://s.example"}}
	if err := UseServer(&cfg); !errors.Is(err, ErrNotSignedIn) {
		t.Fatalf("err = %v, want ErrNotSignedIn", err)
	}
}

func TestRemovedProviderKeyIsRepointed(t *testing.T) {
	path := t.TempDir() + "/c.yml"
	_ = os.WriteFile(path, []byte("provider:\n  backend: openrouter\n"), 0o600)
	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), "Bulle server") {
		t.Fatalf("err = %v, want the server repoint", err)
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
