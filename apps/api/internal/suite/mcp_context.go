package suite

import (
	"context"
	"errors"
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
