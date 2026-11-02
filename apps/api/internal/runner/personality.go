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

// CompilePersonalityPrompt builds system prompt instructions from an assistant profile.
func CompilePersonalityPrompt(profile store.AssistantProfile) string {
	var parts []string
	if profile.Name != "" {
		parts = append(parts, fmt.Sprintf("You are %s, a personal coding assistant. Always introduce and refer to yourself as %s, never as Bulle.", profile.Name, profile.Name))
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

const suitePrompt = "You have direct read access to the Facile suite via courrier_inbox, courrier_search, courrier_accounts, agenda_events, agenda_calendars, opus_tasks, and sablier_status."

func appendSection(base, section string) string {
	if section == "" {
		return base
	}
	if base == "" {
		return section
	}
	return base + "\n\n" + section
}

func (r *Runner) buildSystemPrompt(ctx context.Context, userID int64, channel string) string {
	var profile store.AssistantProfile
	if r.cfg.Store != nil {
		if p, err := r.cfg.Store.GetAssistantProfile(ctx, userID); err == nil {
			profile = p
		}
	}

	system := r.cfg.System
	name := strings.TrimSpace(profile.Name)
	if name != "" && name != "Bulle" && strings.Contains(system, "You are Bulle") {
		replacement := fmt.Sprintf("You are %s, an assistant with its own Linux computer. Always refer to yourself as %s, never as Bulle.", name, name)
		system = strings.Replace(system, "You are Bulle, an assistant with its own Linux computer.", replacement, 1)
	}

	prefix := CompilePersonalityPrompt(profile)
	system = appendSection(prefix, system)
	system = appendSection(system, suitePrompt)
	if channel == store.ChannelWeb {
		system = appendSection(system, r.cfg.WebPromptSuffix)
	}
	return system
}
