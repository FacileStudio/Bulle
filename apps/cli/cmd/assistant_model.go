package cmd

import (
	"context"

	"github.com/FacileStudio/Bulle/apps/cli/internal/agent"
	"github.com/FacileStudio/Bulle/core/settings"
)

type assistantCompleter struct {
	profile string
}

func (c assistantCompleter) Complete(ctx context.Context, prompt string) (string, error) {
	profile, err := agent.AssistantProfile(c.profile)
	if err != nil {
		return "", err
	}
	return agent.RunHeadlessPrompt(ctx, profile, prompt)
}

func assistantSettings() (settings.Assistant, error) {
	agent.EnsureUserPath()
	return agent.AssistantSettings()
}
