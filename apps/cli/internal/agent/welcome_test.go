package agent

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/core/settings"
)

func TestWelcomeDrawsTheAssistantOrKeepsTheConfiguredText(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	config := settings.Defaults("")
	art := *config.StartMessage

	if got := welcome(config, "Jarvis"); strings.Contains(got, "Jarvis") || !strings.Contains(got, "\x1b[38;2;") {
		t.Fatalf("signed in welcome = %q, want the blob in truecolor and no name", got)
	}
	if got := welcome(config, ""); got != art {
		t.Fatalf("no assistant welcome = %q, want the default art", got)
	}
	custom := "hello"
	config.StartMessage = &custom
	if got := welcome(config, "Jarvis"); got != custom {
		t.Fatalf("custom welcome = %q, want the configured text", got)
	}
}

func TestSignedInAssistantAsksTheServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/me" || r.Header.Get("Authorization") != "Bearer good" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"user":{"assistant":{"name":"Jarvis"}}}`))
	}))
	defer srv.Close()
	config := settings.Defaults("")
	config.Server.URL = srv.URL

	t.Setenv("BULLE_TOKEN", "good")
	if got := signedInAssistant(config); got != "Jarvis" {
		t.Fatalf("assistant = %q, want Jarvis", got)
	}
	t.Setenv("BULLE_TOKEN", "bad")
	if got := signedInAssistant(config); got != "" {
		t.Fatalf("rejected token assistant = %q, want none", got)
	}
}
