package suite

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/FacileStudio/nacelle/mcp/client"
)

var errNoUser = errors.New("suite: an MCP request carries no user")

type userKey struct{}

type elicitKey struct{}

// WithUser marks ctx as acting for userID, which is whose delegated token an
// MCP request carries and whose private tool list it reads.
func WithUser(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userKey{}, userID)
}

func userFrom(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userKey{}).(int64)
	return userID, ok
}

// WithElicit hands an MCP call the person who answers its server's
// confirmations. A call without one declines them, so a server that asks
// gets "no" rather than a silent yes.
func WithElicit(ctx context.Context, answer client.Elicit) context.Context {
	return context.WithValue(ctx, elicitKey{}, answer)
}

func elicit(ctx context.Context, request client.Elicitation) (client.Answer, error) {
	answer, ok := ctx.Value(elicitKey{}).(client.Elicit)
	if !ok || answer == nil {
		return client.Answer{}, nil
	}
	return answer(ctx, request)
}

func identity(ctx context.Context) string {
	userID, ok := userFrom(ctx)
	if !ok {
		return ""
	}
	return strconv.FormatInt(userID, 10)
}

// WithConfirmed answers every confirmation of an MCP call with yes, for a
// call the person has just confirmed in Bulle's own interface, such as the
// delete button of a memory page and its confirm dialog.
func WithConfirmed(ctx context.Context) context.Context {
	return WithElicit(ctx, func(_ context.Context, request client.Elicitation) (client.Answer, error) {
		if request.Request == nil {
			return client.Answer{}, nil
		}
		content, err := Acceptance(request.Request.RequestedSchema)
		if err != nil {
			return client.Answer{}, err
		}
		return client.Answer{Accept: true, Content: content}, nil
	})
}

// Acceptance fills the form a server asked to confirm with what an approval
// means: each field's default, and true for a yes-or-no field. A field the
// card cannot answer that way fails the call rather than guess at it.
func Acceptance(schema any) (map[string]any, error) {
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
