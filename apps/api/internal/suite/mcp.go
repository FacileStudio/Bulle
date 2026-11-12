package suite

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/FacileStudio/nacelle"
	"github.com/FacileStudio/nacelle/mcp/client"
)

const (
	// mcpCallLimit bounds one MCP tool call, the user's answer to a
	// confirmation included. It matches how long Courrier keeps a
	// confirmation open, so a card answered later fails on both sides.
	mcpCallLimit = 15 * time.Minute
	// mcpListLimit bounds the handshake and a tools/list.
	mcpListLimit = 20 * time.Second
	// mcpRetryAfter is how long a server that could not be reached is left
	// alone before a run tries to connect again.
	mcpRetryAfter = 30 * time.Second
)

// MCP is Bulle's client of the suite's MCP servers. Each server gets one
// connection, opened lazily under the first user who needs it: the servers
// are stateless, so that connection carries every later user's own token,
// minted per request from the context of the call.
type MCP struct {
	cfg     Config
	tokens  *Tokens
	connect func(ctx context.Context, servers ...client.Server) (*client.Set, error)
	now     func() time.Time

	mu     sync.Mutex
	sets   map[string]*client.Set
	failed map[string]time.Time
	denied map[deniedKey]time.Time
}

func NewMCP(cfg Config, tokens *Tokens) *MCP {
	return &MCP{cfg: cfg, tokens: tokens, connect: client.Connect, now: time.Now, sets: map[string]*client.Set{}, failed: map[string]time.Time{}, denied: map[deniedKey]time.Time{}}
}

// Call runs one of app's MCP tools for userID with args and decodes its
// structured answer into out, which may be nil. A tool error comes back as a
// RefusedError carrying the server's text.
func (m *MCP) Call(ctx context.Context, userID int64, app, name string, args, out any) error {
	if m == nil || !m.cfg.MCP[app].Enabled() {
		return ErrNotConnected
	}
	ctx = WithUser(ctx, userID)
	tools, err := m.appTools(ctx, app)
	if err != nil {
		return err
	}
	index := slices.IndexFunc(tools, func(tool nacelle.Tool) bool { return tool.Spec().Name == name })
	if index < 0 {
		return &AppError{App: app, URL: m.cfg.MCP[app].Site(), Kind: ErrAppDown}
	}
	input, err := json.Marshal(args)
	if err != nil {
		return err
	}
	result, err := tools[index].Run(ctx, nacelle.Call{ID: name, Name: name, Input: input}, func(string) {})
	if err != nil {
		return toolRefusal(app, err)
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal([]byte(result.Text), out); err != nil {
		return fmt.Errorf("%s: reading the answer of %s: %w", app, name, err)
	}
	return nil
}

// toolRefusal turns a tool's error into a RefusedError: 404 when the server
// says the thing named is not there, 400 for any other refusal it explains.
// An outage or a missing account keeps its own error.
func toolRefusal(app string, err error) error {
	var known *AppError
	if errors.As(err, &known) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	text := err.Error()
	status := http.StatusBadRequest
	for _, missing := range []string{"there is no page", "no artifact", "no space with this space_id", "no personal tree"} {
		if strings.Contains(text, missing) {
			status = http.StatusNotFound
		}
	}
	return &RefusedError{App: app, Status: status, Reason: ": " + text}
}

// Probe reports whether userID reaches app's MCP server.
func (m *MCP) Probe(ctx context.Context, userID int64, app string) error {
	_, err := m.appTools(WithUser(ctx, userID), app)
	return err
}

func (m *MCP) appTools(ctx context.Context, app string) ([]nacelle.Tool, error) {
	ctx, cancel := context.WithTimeout(ctx, mcpListLimit)
	defer cancel()
	if err := m.access(ctx, app); err != nil {
		return nil, err
	}
	set, err := m.set(ctx, app)
	if err != nil {
		return nil, m.refusal(app, err)
	}
	listed, err := set.ToolsFor(ctx)
	if err != nil {
		return nil, m.refusal(app, err)
	}
	userID, _ := userFrom(ctx)
	tools := make([]nacelle.Tool, 0, len(listed))
	for _, tool := range listed {
		tools = append(tools, bound(tool, app, userID))
	}
	return tools, nil
}

// set is app's connection, opened under ctx's user the first time. A failed
// connection is retried, but not before mcpRetryAfter.
func (m *MCP) set(ctx context.Context, app string) (*client.Set, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if set := m.sets[app]; set != nil {
		return set, nil
	}
	if at, ok := m.failed[app]; ok && m.now().Sub(at) < mcpRetryAfter {
		return nil, fmt.Errorf("%s did not answer a moment ago; try again shortly", Label(app))
	}
	set, err := m.connect(ctx, m.remote(app))
	if err != nil {
		if _, limited := errors.AsType[*client.RateLimitError](err); !limited && !unauthorized(err) && !forbidden(err) {
			m.failed[app] = m.now()
		}
		return nil, err
	}
	delete(m.failed, app)
	m.sets[app] = set
	return set, nil
}

// refusal names a server that still refuses the user's fresh token as the
// user having no account there: the token is minted for them alone, so the
// server is up and does not know them. A 403 is the server keeping this
// person out. Anything else stays as it is.
func (m *MCP) refusal(app string, err error) error {
	if forbidden(err) {
		return m.noAccess(app)
	}
	if !unauthorized(err) {
		return err
	}
	return &AppError{App: app, URL: m.cfg.MCP[app].Site(), Kind: ErrNoAccount}
}

// unauthorized reports whether err is the 401 the client hands back once a
// refreshed token was refused too. The SDK keeps only the status text.
func unauthorized(err error) bool {
	return strings.Contains(err.Error(), http.StatusText(http.StatusUnauthorized))
}

// remote describes app's server: the user's delegated token on every
// request, the user as the key of private tool lists, the configured host
// allowed even on a private address, and confirmations handed to whoever the
// call's context names.
func (m *MCP) remote(app string) client.Remote {
	server := m.cfg.MCP[app]
	var private []string
	if parsed, err := url.Parse(server.URL); err == nil {
		private = []string{parsed.Hostname()}
	}
	return client.Remote{
		Name:         app,
		URL:          server.URL,
		Timeout:      mcpCallLimit,
		Token:        m.token(app),
		Identity:     identity,
		PrivateHosts: private,
		Elicit:       elicit,
	}
}

func (m *MCP) token(app string) client.TokenSource {
	return func(ctx context.Context) (string, error) {
		userID, ok := userFrom(ctx)
		if !ok {
			return "", errNoUser
		}
		if client.Refreshing(ctx) {
			m.tokens.InvalidateMCP(userID, app)
		}
		return m.tokens.ForMCP(ctx, userID, app)
	}
}

// Close ends every connection.
func (m *MCP) Close() error {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var failures []error
	for app, set := range m.sets {
		failures = append(failures, set.Close())
		delete(m.sets, app)
	}
	return errors.Join(failures...)
}

// bound names a server's tool as the server does, since suite servers
// already prefix their tools with the app, and runs it as userID whatever
// context the loop calls it with.
func bound(tool nacelle.Tool, app string, userID int64) nacelle.Tool {
	return nacelle.Wrap(tool, func(spec nacelle.Spec) nacelle.Spec {
		spec.Name = mcpName(spec.Name, app)
		return spec
	}, func(next nacelle.RunFunc) nacelle.RunFunc {
		return func(ctx context.Context, call nacelle.Call, emit func(string)) (nacelle.Result, error) {
			ctx = WithUser(ctx, userID)
			return throttled(ctx, func() (nacelle.Result, error) { return next(ctx, call, emit) })
		}
	})
}

// throttled sends a call a server refused with 429 once more, after the wait
// the server asked for, when nacelle says the server refused it before it ran.
func throttled(ctx context.Context, run func() (nacelle.Result, error)) (nacelle.Result, error) {
	result, err := run()
	refusal, ok := errors.AsType[*client.RateLimitError](err)
	if !ok || !refusal.Retryable() {
		return result, err
	}
	select {
	case <-ctx.Done():
		return result, err
	case <-time.After(max(refusal.RetryAfter, time.Second)):
	}
	return run()
}

// mcpName drops the server prefix nacelle adds when the tool already carries
// the app's own: courrier_courrier_send becomes courrier_send.
func mcpName(composed, app string) string {
	if trimmed := strings.TrimPrefix(composed, app+"_"); strings.HasPrefix(trimmed, app+"_") {
		return trimmed
	}
	return composed
}
