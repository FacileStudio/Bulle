package hardware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/FacileStudio/nacelle"
)

// CallTimeout bounds one edge tool call, so a dead link never stalls a run.
const CallTimeout = 5 * time.Second

const maxToolName = 64

const maxDescription = 200

type edgeTool struct {
	peer   *Peer
	name   string
	remote string
	desc   string
	schema map[string]any
}

func (t edgeTool) Name() string           { return t.name }
func (t edgeTool) Description() string    { return t.desc }
func (t edgeTool) Schema() map[string]any { return t.schema }

func (t edgeTool) Run(ctx context.Context, input json.RawMessage) (string, error) {
	args := map[string]any{}
	if len(input) > 0 && string(input) != "null" {
		if err := json.Unmarshal(input, &args); err != nil {
			return "", fmt.Errorf("arguments must be a JSON object: %w", err)
		}
	}
	ctx, cancel := context.WithTimeout(ctx, CallTimeout)
	defer cancel()
	result, err := t.peer.Call(ctx, "tools/call", map[string]any{"name": t.remote, "arguments": args})
	if errors.Is(err, context.DeadlineExceeded) {
		return "", fmt.Errorf("%s did not answer within %s", t.peer.Device.Name, CallTimeout)
	}
	if err != nil {
		return "", err
	}
	return resultText(result)
}

// AgentTools turns the user's connected devices into agent tools, and answers
// a short note for the system prompt that names them.
func (r *Registry) AgentTools(userID int64) ([]nacelle.Tool, string) {
	var tools []nacelle.Tool
	var names []string
	taken := map[string]bool{}
	for _, peer := range r.Live(userID) {
		edge, _ := peer.Tools()
		if len(edge) == 0 {
			continue
		}
		names = append(names, peer.Device.Name)
		for _, tool := range edge {
			schema, ok := objectSchema(tool.InputSchema)
			if !ok {
				continue
			}
			name := toolName(peer.Device, tool.Name, taken)
			taken[name] = true
			tools = append(tools, edgeTool{
				peer:   peer,
				name:   name,
				remote: tool.Name,
				desc:   fmt.Sprintf("On the device %q: %s", peer.Device.Name, oneLine(tool.Description, maxDescription)),
				schema: schema,
			})
		}
	}
	if len(names) == 0 {
		return nil, ""
	}
	note := "## Connected devices\n\nThe user's physical BOHS devices are online: " + strings.Join(names, ", ") +
		". When the user asks for something one of them can do, such as switching a light, call that device's tool directly. " +
		"Match the device by its name, even when the user describes it loosely."
	return tools, note
}
