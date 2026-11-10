package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync/atomic"

	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/nacelle"
	"github.com/FacileStudio/nacelle/mcp/client"
)

// elicitKind marks an approval card that shows an MCP server's own question,
// such as Courrier asking before it sends a mail.
const elicitKind = "mcp"

// confirming binds an MCP tool's run to this gate, so a confirmation its
// server asks for mid-call becomes an approval card under the call's id. The
// server owns the decision: Bulle only shows its question and returns the
// owner's answer.
func (g *gate) confirming(tool nacelle.Tool) nacelle.Tool {
	return nacelle.Wrap(tool, nil, func(next nacelle.RunFunc) nacelle.RunFunc {
		return func(ctx context.Context, call nacelle.Call, emit func(string)) (nacelle.Result, error) {
			return next(suite.WithElicit(ctx, g.elicitor(call)), call, emit)
		}
	})
}

// elicitor answers every confirmation one call raises. The first card takes
// the call's own id, so the client shows it on that call; a second question
// in the same call gets a suffix.
func (g *gate) elicitor(call nacelle.Call) client.Elicit {
	var asked atomic.Int32
	return func(ctx context.Context, request client.Elicitation) (client.Answer, error) {
		id := call.ID
		if n := asked.Add(1); n > 1 {
			id += "#" + strconv.Itoa(int(n))
		}
		return g.elicit(ctx, nacelle.Call{ID: id, Name: call.Name, Input: call.Input}, request)
	}
}

// elicit puts the server's question up as an approval card and waits for the
// owner. A denial, a run that ends first, or a card that cannot be recorded
// all decline, and the server then does nothing.
func (g *gate) elicit(ctx context.Context, call nacelle.Call, request client.Elicitation) (client.Answer, error) {
	if request.Request == nil || !g.recorded(ctx, call) {
		return client.Answer{}, nil
	}
	if !g.hold(ctx, call, card{args: elicitCard(request)}) {
		return client.Answer{}, nil
	}
	content, err := acceptance(request.Request.RequestedSchema)
	if err != nil {
		return client.Answer{}, err
	}
	return client.Answer{Accept: true, Content: content}, nil
}

// recorded stores the pending approval of call, or says on the run's stream
// why it could not.
func (g *gate) recorded(ctx context.Context, call nacelle.Call) bool {
	if err := g.store.CreateApproval(context.WithoutCancel(ctx), g.runID, call.ID); err != nil {
		g.unrecorded(call.ID, call.Name, err)
		return false
	}
	g.remember(call.ID, call.Name, call.Input)
	return true
}

// elicitCard is what the owner sees: which server asks, its question, and
// the arguments the model sent the tool.
func elicitCard(request client.Elicitation) map[string]any {
	record := map[string]any{
		"kind":    elicitKind,
		"server":  suite.Label(request.Server),
		"message": request.Request.Message,
	}
	if request.Tool != "" {
		record["tool"] = request.Tool
	}
	if len(request.Arguments) > 0 {
		record["arguments"] = request.Arguments
	}
	if request.Request.URL != "" {
		record["url"] = request.Request.URL
	}
	return record
}

// acceptance fills the form a server asked to confirm with what an approval
// means: each field's default, and true for a yes-or-no field. A field the
// card cannot answer that way fails the call rather than guess at it.
func acceptance(schema any) (map[string]any, error) {
	if schema == nil {
		return nil, nil
	}
	raw, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}
	var form struct {
		Properties map[string]struct {
			Type    string `json:"type"`
			Default any    `json:"default"`
		} `json:"properties"`
		Required []string `json:"required"`
	}
	if err := json.Unmarshal(raw, &form); err != nil {
		return nil, err
	}
	content := map[string]any{}
	for name, field := range form.Properties {
		switch {
		case field.Default != nil:
			content[name] = field.Default
		case field.Type == "boolean":
			content[name] = true
		}
	}
	for _, name := range form.Required {
		if _, ok := content[name]; !ok {
			return nil, fmt.Errorf("the server asked for %q, which an approval card cannot fill in", name)
		}
	}
	return content, nil
}
