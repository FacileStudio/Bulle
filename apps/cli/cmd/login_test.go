package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExchangeCodeAndFetchUser(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/oidc/exchange", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["code"] != "good" {
			http.Error(w, "no", http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"user_id":"1","token":"sess"}`))
	})
	mux.HandleFunc("/api/auth/me", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer sess" {
			http.Error(w, "no", http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"user":{"id":1,"email":"a@b.c","name":"A"}}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	ctx := context.Background()
	if _, err := exchangeCode(ctx, server.URL, "bad"); err == nil {
		t.Fatal("bad code accepted")
	}
	token, err := exchangeCode(ctx, server.URL, "good")
	if err != nil || token != "sess" {
		t.Fatalf("%q %v", token, err)
	}
	user, status, err := fetchUser(ctx, server.URL, token)
	if err != nil || status != 200 || user.Email != "a@b.c" {
		t.Fatalf("%+v %d %v", user, status, err)
	}
	if _, status, _ := fetchUser(ctx, server.URL, "wrong"); status != http.StatusUnauthorized {
		t.Fatalf("status %d", status)
	}
}

func TestPastedLoginUsesNoLoopbackPortAndReadsTheCode(t *testing.T) {
	var out strings.Builder
	code, err := loginCode(context.Background(), &out, strings.NewReader(" abc123 \n"), "https://bulle.example.org", true)
	if err != nil || code != "abc123" {
		t.Fatalf("code %q err %v", code, err)
	}
	if !strings.Contains(out.String(), "https://bulle.example.org/api/auth/oidc?flow=cli\n") || strings.Contains(out.String(), "port=") {
		t.Fatalf("sign-in address = %q, want the cli flow without a port", out.String())
	}
	if _, err := loginCode(context.Background(), &out, strings.NewReader(""), "https://bulle.example.org", true); err == nil {
		t.Fatal("an empty paste was accepted")
	}
}

func TestLoginWithNoURLUsesTheDefaultServer(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home+"/.config")
	t.Setenv("XDG_DATA_HOME", home+"/.local/share")
	t.Setenv("BULLE_SERVER_URL", "")
	got, err := loginServer(nil)
	if err != nil || got != "https://bulle.facile.studio" {
		t.Fatalf("server = %q, %v, want the default", got, err)
	}
	if got, err := loginServer([]string{"https://other.example/"}); err != nil || got != "https://other.example" {
		t.Fatalf("explicit server = %q, %v", got, err)
	}
}
