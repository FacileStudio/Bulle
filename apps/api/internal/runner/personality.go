package runner

import (
	"context"
	"fmt"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

var archetypePrompts = map[string]string{
	"pragmatic_senior":  "Your style is direct, battle-tested, and pragmatic. You prioritize clean, maintainable code over clever tricks. You cut fluff and explain trade-offs clearly.",
	"terse_minimalist":  "Your style is terse and minimalist. You produce code first with minimum prose. You give direct answers without introductory or concluding pleasantries.",
	"supportive_mentor": "Your style is supportive, thorough, and educational. You provide clear, step-by-step explanations and anticipate potential confusion or pitfalls.",
	"witty_hacker":      "Your style is sharp, clever, lightly sarcastic, and highly technical. You appreciate elegant solutions and love deep system insights.",
}

var traitPrompts = map[string]string{
	"concise":           "Keep replies concise and avoid filler words.",
	"explain_reasoning": "Explain the architectural reasoning behind your design decisions.",
	"strict":            "Enforce strict type safety and zero dead code.",
	"creative":          "Suggest creative and innovative approaches when appropriate.",
}

func CompilePersonalityPrompt(profile store.AssistantProfile) string {
	var parts []string
	if profile.Name != "" {
		parts = append(parts, fmt.Sprintf("You are %s, a personal coding assistant.", profile.Name))
	}
	if prompt, ok := archetypePrompts[profile.Archetype]; ok && prompt != "" {
		parts = append(parts, prompt)
	}
	for _, trait := range profile.Traits {
		if prompt, ok := traitPrompts[trait]; ok && prompt != "" {
			parts = append(parts, prompt)
		}
	}
	return strings.Join(parts, "\n\n")
}

func (r *Runner) personalityPrefix(ctx context.Context, userID int64) string {
	if r.cfg.Store == nil {
		return ""
	}
	profile, err := r.cfg.Store.GetAssistantProfile(ctx, userID)
	if err != nil {
		return ""
	}
	return CompilePersonalityPrompt(profile)
}

func (r *Runner) buildSystemPrompt(ctx context.Context, userID int64, channel string) string {
	prefix := r.personalityPrefix(ctx, userID)
	system := r.cfg.System
	if prefix != "" && system != "" {
		system = prefix + "\n\n" + system
	} else if prefix != "" {
		system = prefix
	}
	if channel == store.ChannelWeb && r.cfg.WebPromptSuffix != "" {
		if system != "" {
			system = system + "\n\n" + r.cfg.WebPromptSuffix
		} else {
			system = r.cfg.WebPromptSuffix
		}
	}
	return system
}
