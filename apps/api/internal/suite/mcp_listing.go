package suite

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/FacileStudio/nacelle"
	"github.com/FacileStudio/porte/exchange"
)

// Listing is the MCP tools a run gets and the apps missing from them. Each
// absent app maps to an *AppError whose Kind says why: ErrAppDown when it did
// not answer, which may pass in seconds, ErrNoAccount, ErrNoAccess or
// ErrNotConnected otherwise. Err joins the failures worth a log line.
type Listing struct {
	Tools  []nacelle.Tool
	Absent map[string]*AppError
	Err    error
}

// Down is the absent apps that did not answer, sorted.
func (l Listing) Down() []string {
	var apps []string
	for app, absent := range l.Absent {
		if errors.Is(absent, ErrAppDown) {
			apps = append(apps, app)
		}
	}
	slices.Sort(apps)
	return apps
}

// Tools is every configured MCP server's tools for userID, named as the
// servers name them. An app that cannot be listed is left out and named in
// Absent, so one app being down never costs the user the rest, and the run
// can tell the person why. An app the user may not use is not logged.
func (m *MCP) Tools(ctx context.Context, userID int64) Listing {
	if m == nil {
		return Listing{Absent: map[string]*AppError{}}
	}
	return m.ToolsOf(ctx, userID, m.cfg.EnabledMCP())
}

// ToolsOf is Tools for the named apps only, such as the ones a run could not
// list when it started.
func (m *MCP) ToolsOf(ctx context.Context, userID int64, apps []string) Listing {
	out := Listing{Absent: map[string]*AppError{}}
	if m == nil {
		return out
	}
	ctx = WithUser(ctx, userID)
	var failures []error
	for _, app := range apps {
		if facadeMCP[app] || !m.cfg.MCP[app].Enabled() {
			continue
		}
		found, err := m.appTools(ctx, app)
		if err == nil {
			out.Tools = append(out.Tools, found...)
			continue
		}
		out.Absent[app] = &AppError{App: app, URL: m.cfg.MCP[app].Site(), Kind: absence(err)}
		if !errors.Is(err, ErrNoAccess) {
			failures = append(failures, fmt.Errorf("suite: %s MCP: %w", app, err))
		}
	}
	out.Err = errors.Join(failures...)
	return out
}

// absence is why an app's tools could not be listed. Registre refusing the
// exchange for good is the person having no access through Bulle; anything
// else that is not a known refusal is the app or Registre not answering.
func absence(err error) error {
	for _, kind := range []error{ErrNoAccess, ErrNoAccount, ErrNotConnected} {
		if errors.Is(err, kind) {
			return kind
		}
	}
	var refused *exchange.RefusedError
	if errors.As(err, &refused) && !transientExchange(err) {
		return ErrNoAccess
	}
	return ErrAppDown
}
