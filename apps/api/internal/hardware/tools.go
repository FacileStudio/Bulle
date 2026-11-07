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
	remote string
	spec   nacelle.Spec
}

func (t edgeTool) Spec() nacelle.Spec { return t.spec }

func (t edgeTool) Run(ctx context.Context, call nacelle.Call, _ func(string)) (nacelle.Result, error) {
	args := map[string]any{}
	if len(call.Input) > 0 && string(call.Input) != "null" {
		if err := json.Unmarshal(call.Input, &args); err != nil {
			return nacelle.Result{}, fmt.Errorf("arguments must be a JSON object: %w", err)
		}
	}
	ctx, cancel := context.WithTimeout(ctx, CallTimeout)
	defer cancel()
	result, err := t.peer.Call(ctx, "tools/call", map[string]any{"name": t.remote, "arguments": args})
	if errors.Is(err, context.DeadlineExceeded) {
		return nacelle.Result{}, fmt.Errorf("%s did not answer within %s", t.peer.Device.Name, CallTimeout)
	}
	if err != nil {
		return nacelle.Result{}, err
	}
	text, err := resultText(result)
	return nacelle.Result{Text: text}, err
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
			desc := fmt.Sprintf("On the device %q: %s", peer.Device.Name, oneLine(tool.Description, maxDescription))
			spec := nacelle.Spec{Name: name, Description: desc, Schema: schema}
			tools = append(tools, edgeTool{peer: peer, remote: tool.Name, spec: spec})
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
