package provider

import (
	"testing"
)

func TestNewAnthropic(t *testing.T) {
	b, err := New(Config{Backend: "anthropic", Model: "claude-opus-5"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b == nil {
		t.Fatal("expected non-nil backend")
	}
}

func TestNewAnthropicRejectsKeyWithoutEndpoint(t *testing.T) {
	_, err := New(Config{Backend: "anthropic", APIKey: "k"})
	if err == nil {
		t.Fatal("expected error when api_key has no base_url")
	}
}

func TestNewAnthropicAcceptsEndpoint(t *testing.T) {
	if _, err := New(Config{Backend: "anthropic", BaseURL: "http://localhost:8080", APIKey: "tok"}); err != nil {
		t.Fatal(err)
	}
}

func TestNewOpenRouterRequiresModel(t *testing.T) {
	_, err := New(Config{Backend: "openrouter"})
	if err == nil {
		t.Fatal("expected error when model is missing for openrouter")
	}
}

func TestNewUnknownBackend(t *testing.T) {
	_, err := New(Config{Backend: "unknown"})
	if err == nil {
		t.Fatal("expected error for unknown backend")
	}
}
