package suite

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/FacileStudio/porte/exchange"
)

// mcpDeniedFor is how long Bulle remembers that Registre refused a person an
// app, so their runs and status checks stop asking until an administrator may
// have granted it.
const mcpDeniedFor = 5 * time.Minute

type deniedKey struct {
	user int64
	app  string
}

// access mints the user's token for app before any connection is touched. A
// person Registre does not let use the app, such as a member on an
// admin-only app, gets ErrNoAccess and no connection, so they never mark the
// server down for everyone else.
func (m *MCP) access(ctx context.Context, app string) error {
	userID, ok := userFrom(ctx)
	if !ok {
		return errNoUser
	}
	key := deniedKey{user: userID, app: app}
	if m.deniedAt(key) {
		return m.noAccess(app)
	}
	_, err := m.tokens.ForMCP(ctx, userID, app)
	m.mu.Lock()
	defer m.mu.Unlock()
	if !refusedAccess(err) {
		delete(m.denied, key)
		return err
	}
	m.denied[key] = m.now()
	return m.noAccess(app)
}

func (m *MCP) deniedAt(key deniedKey) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	at, ok := m.denied[key]
	return ok && m.now().Sub(at) < mcpDeniedFor
}

func (m *MCP) noAccess(app string) error {
	return &AppError{App: app, URL: m.cfg.MCP[app].Site(), Kind: ErrNoAccess}
}

// refusedAccess reports whether Registre refused the exchange because the
// person may not use the app.
func refusedAccess(err error) bool {
	var refused *exchange.RefusedError
	return errors.As(err, &refused) && refused.Code == "access_denied"
}

// forbidden reports whether err is a server's 403: the token is valid, but
// the server does not let this person in. The SDK keeps only the status text.
func forbidden(err error) bool {
	return strings.Contains(err.Error(), http.StatusText(http.StatusForbidden))
}
