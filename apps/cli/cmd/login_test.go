package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
