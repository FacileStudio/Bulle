package main

import (
	"context"

	"github.com/FacileStudio/Bulle/apps/api/internal/catalog"
	"github.com/FacileStudio/Bulle/apps/api/internal/env"
	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	aimodule "github.com/FacileStudio/Bulle/apps/api/modules/ai"
	llmmodule "github.com/FacileStudio/Bulle/apps/api/modules/llm"
)

// llmOffer is what GET /api/llm hands a client: the instance's backend and
// models, the backend's full catalogue, the user's own default model when they
// saved one, and the user's assistant persona compiled the way the web runner
// compiles it, so the CLI and the web app answer as the same assistant.
func llmOffer(cfg env.LLM, st store.Store, catalog func(ctx context.Context, userID int64) []catalog.Model) llmmodule.Offer {
	offer := llmmodule.Offer{
		Config:  llmmodule.Config{Backend: cfg.Backend, Model: cfg.Model, Models: cfg.Models},
		Catalog: catalog,
	}
	if st == nil {
		return offer
	}
	offer.ModelFor = func(ctx context.Context, userID int64) string {
		prefs, err := st.GetPreferences(ctx, userID)
		if err != nil {
			return ""
		}
		return prefs.Model
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

func summarizer(rn *runner.Runner) aimodule.Summarizer {
	if rn == nil {
		return nil
	}
	return rn
}

func completer(rn *runner.Runner) aimodule.Completer {
	if rn == nil {
		return nil
	}
	return rn
}
