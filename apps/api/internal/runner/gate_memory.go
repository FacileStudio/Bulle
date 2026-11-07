package runner

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

// memoryPolicy shows the owner which memory page a memory_delete removes and
// in which space, never the model's raw arguments. It takes no grant, so
// every delete of a tainted run is asked on its own.
type memoryPolicy struct {
	space string
}

func (p memoryPolicy) describe(_ context.Context, input json.RawMessage) (map[string]any, string, error) {
	var in struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, "", err
	}
	path := strings.TrimSpace(in.Path)
	if path == "" {
		return nil, "", errors.New("memory_delete names no page")
	}
	record := map[string]any{"kind": "memory", "action": "delete", "path": path}
	if p.space != "" {
		record["space"] = p.space
	}
	return record, "", nil
}

func (p memoryPolicy) denied(context.Context, json.RawMessage) {}
