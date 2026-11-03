package agent

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/core/settings"
)

func TestWelcomeDrawsTheServerAssistantOrKeepsTheConfiguredText(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/me" || r.Header.Get("Authorization") != "Bearer good" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if _, err := w.Write([]byte(`{"user":{"assistant":{"name":"Jarvis"}}}`)); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()
	t.Setenv("NO_COLOR", "")

	config := settings.Defaults("")
	config.Server.URL = srv.URL
	art := *config.StartMessage

	t.Setenv("BULLE_TOKEN", "good")
	if got := welcome(config); !strings.Contains(got, "Jarvis") || !strings.Contains(got, "\x1b[38;2;") {
		t.Fatalf("signed in welcome = %q, want Jarvis drawn in truecolor", got)
	}
	t.Setenv("BULLE_TOKEN", "bad")
	if got := welcome(config); got != art {
		t.Fatalf("rejected token welcome = %q, want the default art", got)
	}
	custom := "hello"
	config.StartMessage = &custom
	t.Setenv("BULLE_TOKEN", "good")
	if got := welcome(config); got != custom {
		t.Fatalf("custom welcome = %q, want the configured text", got)
	}
}
