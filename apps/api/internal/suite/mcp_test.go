package suite

import (
	"context"
	"errors"
	"testing"

	"github.com/FacileStudio/nacelle"
	"github.com/FacileStudio/nacelle/mcp/client"
)

func TestThrottledSendsAGivenUpCallOnceMore(t *testing.T) {
	calls := 0
	refused := &client.RateLimitError{Server: "glouton", Method: "tools/call", Status: 429}
	result, err := throttled(context.Background(), func() (nacelle.Result, error) {
		calls++
		if calls == 1 {
			return nacelle.Result{}, refused
		}
		return nacelle.Result{Text: "ok"}, nil
	})
	if err != nil || result.Text != "ok" || calls != 2 {
		t.Fatalf("a first call refused with 429 is sent once more: calls=%d, %v", calls, err)
	}
	calls = 0
	retried := &client.RateLimitError{Server: "glouton", Method: "tools/list", Status: 429, Retries: 3}
	if _, err := throttled(context.Background(), func() (nacelle.Result, error) { calls++; return nacelle.Result{}, retried }); !errors.Is(err, retried) || calls != 1 {
		t.Fatalf("a refusal nacelle already retried is not sent again: calls=%d, %v", calls, err)
	}
}

func TestARateLimitAtConnectDoesNotMarkTheServerDown(t *testing.T) {
	m := NewMCP(Config{MCP: map[string]MCPConfig{AppGlouton: {URL: "https://glouton.test/mcp", Resource: "https://glouton.test/mcp"}}}, nil)
	m.connect = func(context.Context, ...client.Server) (*client.Set, error) {
		return nil, &client.RateLimitError{Server: AppGlouton, Method: "initialize", Status: 429}
	}
	if _, err := m.set(context.Background(), AppGlouton); err == nil {
		t.Fatal("the refusal must come back")
	}
	if len(m.failed) != 0 {
		t.Fatalf("a 429 means the server is up: %v", m.failed)
	}
}
