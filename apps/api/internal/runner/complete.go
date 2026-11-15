package runner

import (
	"context"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/aiusage"
	"github.com/FacileStudio/nacelle"
)

const completeMaxTokens = 1024

// Complete answers one prompt with the instance's default model and no tool,
// for the service routes that need plain text written by the user's AI. It
// does not check ai.access: its callers do.
func (r *Runner) Complete(ctx context.Context, userID int64, system, prompt string) (string, aiusage.Call, error) {
	return r.CompleteLimit(ctx, userID, system, prompt, completeMaxTokens)
}

// CompleteLimit is Complete with the caller's output token budget. The Call it
// returns carries the usage the backend reported, also when the provider
// failed after reporting some.
func (r *Runner) CompleteLimit(ctx context.Context, userID int64, system, prompt string, maxTokens int64) (string, aiusage.Call, error) {
	name, model := r.Defaults()
	call := aiusage.Call{Model: model}
	backend, own, err := r.backend(ctx, userID, name, model)
	if err != nil {
		return "", call, err
	}
	call.OwnKey = own
	if !own && r.cfg.Budget != nil {
		if err := r.cfg.Budget.Check(ctx, userID); err != nil {
			return "", call, err
		}
	}
	request := nacelle.Request{System: system, Messages: []nacelle.Message{nacelle.UserText(prompt)}, MaxTokens: maxTokens, MaxIterations: 1}
	var text strings.Builder
	for event, err := range backend.Stream(ctx, request) {
		if err != nil {
			return "", call, err
		}
		switch event.Kind {
		case nacelle.KindText:
			text.WriteString(event.Text)
		case nacelle.KindTurn:
			call.Usage = call.Usage.Add(event.Usage)
		case nacelle.KindDone:
			call.Usage = event.Usage
		}
	}
	return text.String(), call, nil
}
