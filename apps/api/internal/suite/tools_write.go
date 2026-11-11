package suite

import (
	"context"

	"github.com/FacileStudio/nacelle"
)

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

// SelfWrites are the tools that skip the approval card on purpose: the
// agent's own notes and artifacts, which the user edits or deletes in
// Mycelium.
func SelfWrites() []string {
	return []string{"memory_write", "memory_delete", "artifact_create", "artifact_update"}
}
