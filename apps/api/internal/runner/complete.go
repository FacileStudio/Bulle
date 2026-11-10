package runner

import (
	"context"
	"strings"

	"github.com/FacileStudio/nacelle"
)

const completeMaxTokens = 1024

// Complete answers one prompt with the instance's default model and no tool,
// for the service routes that need plain text written by the user's AI. It
// does not check ai.access: its callers do.
func (r *Runner) Complete(ctx context.Context, userID int64, system, prompt string) (string, string, error) {
	name, model := r.Defaults()
	backend, err := r.backend(ctx, userID, name, model)
	if err != nil {
		return "", model, err
	}
	request := nacelle.Request{System: system, Messages: []nacelle.Message{nacelle.UserText(prompt)}, MaxTokens: completeMaxTokens, MaxIterations: 1}
	var text strings.Builder
	for event, err := range backend.Stream(ctx, request) {
		if err != nil {
			return "", model, err
		}
		if event.Kind == nacelle.KindText {
			text.WriteString(event.Text)
		}
	}
	return text.String(), model, nil
}
