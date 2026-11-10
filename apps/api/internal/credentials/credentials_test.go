package credentials

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

func testKey() []byte { return bytes.Repeat([]byte{7}, 32) }

func TestRoundTripAndAADBinding(t *testing.T) {
	for _, name := range []string{"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GEMINI_API_KEY", "GOOGLE_API_KEY", "OPENROUTER_API_KEY"} {
		t.Setenv(name, "")
	}
	ctx := context.Background()
	svc, err := NewMemory(testKey())
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Put(ctx, 1, "anthropic", "sk-ant-secret-1234"); err != nil {
		t.Fatal(err)
	}
	if err := svc.Put(ctx, 1, "openai", "sk-openai-secret-5678"); err != nil {
		t.Fatal(err)
	}
	got, err := svc.Resolve(ctx, 1, "anthropic")
	if err != nil || got != "sk-ant-secret-1234" {
		t.Fatalf("round trip: %q %v", got, err)
	}
	entries, _ := svc.List(ctx, 1)
	if len(entries) != 2 || entries[0].Last4 != "1234" {
		t.Fatalf("entries: %+v", entries)
	}
	if _, err := svc.Resolve(ctx, 2, "anthropic"); !errors.Is(err, ErrNoKey) {
		t.Fatalf("other user must not resolve: %v", err)
	}

	svc.repo.(*memory).swap(1, "anthropic", "openai")
	if _, err := svc.Resolve(ctx, 1, "anthropic"); err == nil || errors.Is(err, ErrNoKey) {
		t.Fatalf("swapped row must fail to decrypt: %v", err)
	}

	m := svc.repo.(*memory)
	m.rows[memoryKey(2, "anthropic")] = m.rows[memoryKey(1, "openai")]
	if _, err := svc.Resolve(ctx, 2, "anthropic"); err == nil {
		t.Fatal("row moved to another user must fail to decrypt")
	}
}

func TestResolveFallsBackToInstanceKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-instance")
	svc, _ := NewMemory(nil)
	got, err := svc.Resolve(context.Background(), 9, "openai")
	if err != nil || got != "sk-instance" {
		t.Fatalf("fallback: %q %v", got, err)
	}
	if err := svc.Put(context.Background(), 9, "openai", "x"); !errors.Is(err, ErrDisabled) {
		t.Fatalf("put without master key: %v", err)
	}
}
