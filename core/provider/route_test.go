package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	sdk "github.com/anthropics/anthropic-sdk-go"
)

func TestAnthropicRoutesThroughBaseURLWithBearer(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "vendor-secret")
	var path, auth, apiKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, auth, apiKey = r.URL.Path, r.Header.Get("Authorization"), r.Header.Get("X-Api-Key")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := sdk.NewClient(anthropicOptions(Config{BaseURL: server.URL + "/llm/anthropic", APIKey: "bulle-tok"})...)
	_, _ = client.Messages.New(context.Background(), sdk.MessageNewParams{Model: "m", MaxTokens: 1})

	if path != "/llm/anthropic/v1/messages" {
		t.Fatalf("path %q", path)
	}
	if auth != "Bearer bulle-tok" {
		t.Fatalf("authorization %q", auth)
	}
	if apiKey != "" {
		t.Fatalf("vendor key leaked: %q", apiKey)
	}
}
