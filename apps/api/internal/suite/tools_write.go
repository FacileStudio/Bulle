package suite

import (
	"context"

	"github.com/FacileStudio/nacelle"
)

func writeTool[In any](set *toolSet, name, description string, run func(ctx context.Context, in In) (string, error)) {
	tool, err := nacelle.NewTool(name, description, run)
	set.add(tool, err)
	set.writes = append(set.writes, name)
}

func done(message string, err error) (string, error) {
	if err != nil {
		return "", err
	}
	return message, nil
}

func selfTool[In any](set *toolSet, name, description string, run func(ctx context.Context, in In) (string, error)) {
	tool, err := nacelle.NewTool(name, description, run)
	set.add(tool, err)
}

// SelfWrites are the tools that change only the agent's own notes. They skip
// the approval card on purpose: the user sees them in the tool list and edits
// the notes in Mycelium.
func SelfWrites() []string { return []string{"memory_write"} }
