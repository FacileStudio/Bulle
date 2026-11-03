package cmd

import (
	"context"

	"github.com/FacileStudio/Bulle/apps/cli/internal/agent"
	"github.com/FacileStudio/Bulle/core/settings"
)

type assistantCompleter struct {
	model string
}

func (c assistantCompleter) Complete(ctx context.Context, prompt string) (string, error) {
	model, err := agent.AssistantModel(c.model)
	if err != nil {
		return "", err
	}
	return agent.RunHeadlessPrompt(ctx, model, prompt)
}

func assistantSettings() (settings.Assistant, error) {
	agent.EnsureUserPath()
	return agent.AssistantSettings()
}
