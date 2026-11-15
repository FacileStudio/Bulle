package suite

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func openCapsuleText(t *testing.T, ciphertext, key string) string {
	t.Helper()
	raw, err := base64.RawURLEncoding.DecodeString(key)
	if err != nil || len(raw) != capsuleKeyBytes {
		t.Fatalf("key is not 32 bytes of base64url: %v", err)
	}
	combined, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		t.Fatalf("ciphertext is not standard base64: %v", err)
	}
	aead, err := capsuleAEAD(raw)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := aead.Open(nil, combined[:capsuleIVBytes], combined[capsuleIVBytes:], nil)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	return string(plain)
}

func TestSealCapsuleTextMatchesCapsuleFormat(t *testing.T) {
	ciphertext, key, err := sealCapsuleText("the wifi password is hunter2")
	if err != nil {
		t.Fatal(err)
	}
	if strings.ContainsAny(key, "+/=") {
		t.Fatalf("key %q must be base64url without padding", key)
	}
	if got := openCapsuleText(t, ciphertext, key); got != "the wifi password is hunter2" {
		t.Fatalf("round trip = %q", got)
	}
}

func TestSealerStoresOnlyCiphertext(t *testing.T) {
	var got map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/pastes" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"cap_0123456789abcdef","delete_token":"tok","created_at":"2026-11-16T00:00:00Z"}`))
	}))
	defer server.Close()
	sealer := NewSealer(MCPConfig{URL: server.URL + "/mcp", Resource: "https://capsule.example/mcp"}, server.Client())
	sealed, err := sealer.Seal(context.Background(), "secret text", "7d", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got["expires_in"] != "7d" || got["burn_after_read"] != true {
		t.Fatalf("Capsule got fields %v, want only content, burn_after_read and expires_in", got)
	}
	content, _ := got["content"].(string)
	if strings.Contains(content, "secret") {
		t.Fatal("Capsule received plaintext")
	}
	prefix := "https://capsule.example/cap_0123456789abcdef#"
	if !strings.HasPrefix(sealed.URL, prefix) || sealed.DeleteToken != "tok" {
		t.Fatalf("sealed = %+v", sealed)
	}
	if plain := openCapsuleText(t, content, strings.TrimPrefix(sealed.URL, prefix)); plain != "secret text" {
		t.Fatalf("link key opens %q", plain)
	}
}

func TestSealerRefusesBadInput(t *testing.T) {
	sealer := NewSealer(MCPConfig{URL: "http://127.0.0.1:1/mcp", Resource: "https://capsule.example/mcp"}, nil)
	for _, c := range []struct{ text, expires string }{{"  ", "24h"}, {"x", "2h"}, {strings.Repeat("a", maxCapsuleText+1), "1h"}} {
		if _, err := sealer.Seal(context.Background(), c.text, c.expires, true); err == nil {
			t.Fatalf("Seal(%q, %q) accepted bad input", c.text[:min(len(c.text), 8)], c.expires)
		}
	}
	if NewSealer(MCPConfig{}, nil) != nil {
		t.Fatal("an unconfigured Capsule must give no sealer")
	}
}
