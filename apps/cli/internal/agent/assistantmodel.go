package agent

import (
	"context"

	"github.com/FacileStudio/Bulle/core/settings"
)

// AssistantModel returns the server model the assistant calls use: the given
// name, else the chat model from the config, else the server's default.
func AssistantModel(name string) (string, error) {
	if name != "" {
		return name, nil
	}
	cfg, err := loadCronConfig()
	if err != nil {
		return "", err
	}
	return cfg.Chat.Model, nil
}

// RunHeadlessPrompt runs one prompt through the headless agent on the named
// server model with no tool, MCP server or hook, and returns the text. Cron
// jobs resolve config the same way.
func RunHeadlessPrompt(ctx context.Context, model, prompt string) (string, error) {
	cfg, err := loadCronConfig()
	if err != nil {
		return "", err
	}
	cfg.Model = model
	return runToolless(ctx, cfg, prompt)
}

// AssistantSettings returns the assistant block of the resolved config.
func AssistantSettings() (settings.Assistant, error) {
	cfg, err := loadCronConfig()
	return cfg.Assistant, err
}
