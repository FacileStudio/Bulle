package runner

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

const askUserTool = "ask_user"

const askUserDescription = "Ask the user one to four multiple-choice questions on a card, when you cannot go on without a choice only they can make: " +
	"an ambiguous request, two valid approaches with a real trade-off, or a missing detail you cannot find with your tools. " +
	"Do not use it for a simple yes or no, for something you can look up, or to confirm an action that already shows an approval card. " +
	"Give each question two to four distinct options, the one you recommend first. The card always adds a field where the user types their own answer, so never add an Other option. " +
	"After calling it, end your turn without repeating the questions: the answers arrive as the user's next message."

type askOption struct {
	Label       string `json:"label" jsonschema:"required,description=The choice in one to five words"`
	Description string `json:"description,omitempty" jsonschema:"description=What picking it means or costs, in one short sentence"`
}

type askQuestion struct {
	Question    string      `json:"question" jsonschema:"required,description=The full question, ending with a question mark"`
	Header      string      `json:"header,omitempty" jsonschema:"description=A label of at most 12 characters, such as Date or Tone"`
	Options     []askOption `json:"options" jsonschema:"required,description=Two to four distinct choices, the recommended one first"`
	MultiSelect bool        `json:"multi_select,omitempty" jsonschema:"description=True when the user may pick several options"`
}

type askUserInput struct {
	Questions []askQuestion `json:"questions" jsonschema:"required,description=One to four questions"`
}

// askUserTools gives a web chat the ask_user card. Other channels have no
// card to draw it on, so their runs ask in prose.
func askUserTools(channel string) ([]nacelle.Tool, error) {
	if channel != store.ChannelWeb && channel != "" {
		return nil, nil
	}
	tool, err := nacelle.NewToolWithOptions(askUserTool, askUserDescription, func(_ context.Context, in askUserInput) (string, error) {
		if err := in.validate(); err != nil {
			return "", err
		}
		return "The user sees your questions on a card now. End your turn here, without repeating them; their answers arrive as their next message.", nil
	}, nacelle.ToolOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	return []nacelle.Tool{tool}, nil
}

func (in askUserInput) validate() error {
	if len(in.Questions) == 0 || len(in.Questions) > 4 {
		return errors.New("ask one to four questions")
	}
	for i, q := range in.Questions {
		if strings.TrimSpace(q.Question) == "" {
			return fmt.Errorf("question %d is empty", i+1)
		}
		if len(q.Options) < 2 || len(q.Options) > 4 {
			return fmt.Errorf("question %d needs two to four options", i+1)
		}
		labels := map[string]bool{}
		for _, option := range q.Options {
			label := strings.ToLower(strings.TrimSpace(option.Label))
			if label == "" || labels[label] {
				return fmt.Errorf("question %d has an empty or repeated option", i+1)
			}
			if label == "other" {
				return fmt.Errorf("question %d: drop the Other option, the card adds its own", i+1)
			}
			labels[label] = true
		}
	}
	return nil
}
