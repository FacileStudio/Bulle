package runner

import (
	"context"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/core/askuser"
	"github.com/FacileStudio/nacelle"
)

const askUserDescription = askuser.Description + " The questions show on a card. " +
	"After calling it, end your turn without repeating them: the answers arrive as the user's next message."

// askUserTools gives a web chat the ask_user card. Other channels have no
// card to draw it on, so their runs ask in prose.
func askUserTools(channel string) ([]nacelle.Tool, error) {
	if channel != store.ChannelWeb && channel != "" {
		return nil, nil
	}
	tool, err := nacelle.NewToolWithOptions(askuser.Name, askUserDescription, func(_ context.Context, in askuser.Input) (string, error) {
		if err := in.Validate(); err != nil {
			return "", err
		}
		return "The user sees your questions on a card now. End your turn here, without repeating them; their answers arrive as their next message.", nil
	}, nacelle.ToolOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	return []nacelle.Tool{tool}, nil
}
