package runner

import (
	"context"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

func TestCompilePersonalityPromptEmpty(t *testing.T) {
	if got := CompilePersonalityPrompt(store.AssistantProfile{}); got != "" {
		t.Fatalf("expected empty prompt, got %q", got)
	}
}

func TestCompilePersonalityPromptNameOnly(t *testing.T) {
	profile := store.AssistantProfile{Name: "Bulle"}
	want := "You are Bulle, a personal coding assistant. Always introduce and refer to yourself as Bulle, never as Bulle."
	if got := CompilePersonalityPrompt(profile); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestCompilePersonalityPromptArchetypes(t *testing.T) {
	cases := []struct {
		archetype string
		contains  string
	}{
		{"pragmatic_senior", "Your style is direct, battle-tested, and pragmatic."},
		{"terse_minimalist", "Your style is terse and minimalist."},
		{"supportive_mentor", "Your style is supportive, thorough, and educational."},
		{"witty_hacker", "Your style is sharp, clever, lightly sarcastic, and highly technical."},
	}
	for _, tc := range cases {
		profile := store.AssistantProfile{Archetype: tc.archetype}
		got := CompilePersonalityPrompt(profile)
		if !strings.Contains(got, tc.contains) {
			t.Fatalf("archetype %s: expected %q in %q", tc.archetype, tc.contains, got)
		}
	}
}

func TestCompilePersonalityPromptTraits(t *testing.T) {
	profile := store.AssistantProfile{
		Traits: []string{"concise", "explain_reasoning", "strict", "creative"},
	}
	got := CompilePersonalityPrompt(profile)
	for _, snippet := range []string{
		"Keep replies concise and avoid filler words.",
		"Explain the architectural reasoning behind your design decisions.",
		"Enforce strict type safety and zero dead code.",
		"Suggest creative and innovative approaches when appropriate.",
	} {
		if !strings.Contains(got, snippet) {
			t.Fatalf("expected %q in %q", snippet, got)
		}
	}
}

func TestCompilePersonalityPromptFull(t *testing.T) {
	profile := store.AssistantProfile{
		Name:      "Atlas",
		Archetype: "witty_hacker",
		Traits:    []string{"concise", "strict"},
	}
	got := CompilePersonalityPrompt(profile)
	want := "You are Atlas, a personal coding assistant. Always introduce and refer to yourself as Atlas, never as Bulle.\n\nYour style is sharp, clever, lightly sarcastic, and highly technical. You appreciate elegant solutions and love deep system insights.\n\nKeep replies concise and avoid filler words.\n\nEnforce strict type safety and zero dead code."
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestBuildSystemPrompt(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	_ = st.SaveAssistantProfile(ctx, store.AssistantProfile{
		UserID:    42,
		Name:      "Atlas",
		Archetype: "pragmatic_senior",
	})
	r := &Runner{
		cfg: Config{
			System:          "You are Bulle, an assistant with its own Linux computer. Base prompt.",
			WebPromptSuffix: "Web suffix.",
			Store:           st,
		},
	}

	withProfile := r.buildSystemPrompt(ctx, 42, store.ChannelWeb, "Tools line.")
	if !strings.HasPrefix(withProfile, "You are Atlas, a personal coding assistant.") {
		t.Fatalf("expected personality prefix, got %q", withProfile)
	}
	if !strings.Contains(withProfile, "You are Atlas, an assistant with its own Linux computer. Always refer to yourself as Atlas, never as Bulle. Base prompt.") {
		t.Fatalf("expected replaced base prompt, got %q", withProfile)
	}
	if strings.Contains(withProfile, "You are Bulle") {
		t.Fatalf("expected no mention of Bulle, got %q", withProfile)
	}
	if !strings.Contains(withProfile, "Tools line.") {
		t.Fatalf("expected the extra section, got %q", withProfile)
	}
	if !strings.HasSuffix(withProfile, "Web suffix.") {
		t.Fatalf("expected web suffix, got %q", withProfile)
	}

	withoutProfile := r.buildSystemPrompt(ctx, 999, "", "")
	wantWithout := "You are Bulle, an assistant with its own Linux computer. Base prompt."
	if withoutProfile != wantWithout {
		t.Fatalf("expected unchanged base prompt, got %q", withoutProfile)
	}
}
