package main

import (
	"context"

	"github.com/FacileStudio/Bulle/apps/api/internal/env"
	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	llmmodule "github.com/FacileStudio/Bulle/apps/api/modules/llm"
)

// llmOffer is what GET /api/llm hands a client: the instance's backend and
// models, and the user's assistant persona compiled the way the web runner
// compiles it, so the CLI and the web app answer as the same assistant.
func llmOffer(cfg env.LLM, st store.Store) llmmodule.Offer {
	offer := llmmodule.Offer{Config: llmmodule.Config{Backend: cfg.Backend, Model: cfg.Model, Models: cfg.Models}}
	if st == nil {
		return offer
	}
	offer.Persona = func(ctx context.Context, userID int64) string {
		profile, err := st.GetAssistantProfile(ctx, userID)
		if err != nil {
			return ""
		}
		return runner.CompilePersonalityPrompt(profile)
	}
	return offer
}
