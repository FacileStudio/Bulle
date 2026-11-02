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
