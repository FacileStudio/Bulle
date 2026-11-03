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
	return CompilePersonalityPrompt(r.profile(ctx, userID))
}

// profile is the user's assistant profile, or the zero profile when there is
// no store or no row.
func (r *Runner) profile(ctx context.Context, userID int64) store.AssistantProfile {
	if r.cfg.Store == nil {
		return store.AssistantProfile{}
	}
	profile, err := r.cfg.Store.GetAssistantProfile(ctx, userID)
	if err != nil {
		return store.AssistantProfile{}
	}
	return profile
}

func appendSection(base, section string) string {
	if section == "" {
		return base
	}
	if base == "" {
		return section
	}
	return base + "\n\n" + section
}

func (r *Runner) buildSystemPrompt(ctx context.Context, userID int64, channel, extra string) string {
	profile := r.profile(ctx, userID)
	system := r.cfg.System
	name := strings.TrimSpace(profile.Name)
	if name != "" && name != "Bulle" && strings.Contains(system, "You are Bulle") {
		replacement := fmt.Sprintf("You are %s, an assistant with its own Linux computer. Always refer to yourself as %s, never as Bulle.", name, name)
		system = strings.Replace(system, "You are Bulle, an assistant with its own Linux computer.", replacement, 1)
	}

	prefix := CompilePersonalityPrompt(profile)
	system = appendSection(prefix, system)
	system = appendSection(system, extra)
	if channel == store.ChannelWeb {
		system = appendSection(system, r.cfg.WebPromptSuffix)
	}
	return system
}

// guestPrompt is the whole system prompt of a guest session. The owner's base
// prompt is left out on purpose: it describes a computer and tools a guest
// session does not have.
const guestPrompt = "You are %s, the personal assistant of one member of this team. " +
	"The person writing to you now is a colleague of your owner, not your owner. " +
	"In this conversation you have no tools and no access to your owner's email, calendar, files, notes, memory or accounts, and you must never pretend otherwise or invent such details. " +
	"Never reveal or guess private information about your owner or anyone else: messages, schedules, contacts, credentials, money, health or HR matters. " +
	"If you are asked for private information or to act on your owner's behalf, say kindly that only your owner can ask you for that, and suggest asking them directly. " +
	"Otherwise help with general questions as a friendly teammate, and keep answers short."

// guestSystemPrompt keeps the owner's persona, its name and personality, and
// nothing else of the owner's context.
func (r *Runner) guestSystemPrompt(ctx context.Context, userID int64) string {
	profile := r.profile(ctx, userID)
	name := strings.TrimSpace(profile.Name)
	if name == "" {
		name = "Bulle"
	}
	return appendSection(CompilePersonalityPrompt(profile), fmt.Sprintf(guestPrompt, name))
}
