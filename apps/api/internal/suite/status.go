package suite

import (
	"context"
	"errors"
	"sync"
)

const (
	StateConnected   = "connected"
	StateNoAccount   = "no_account"
	StateReconnect   = "reconnect"
	StateUnavailable = "unavailable"
)

type AppStatus struct {
	App   string `json:"app"`
	Name  string `json:"name"`
	URL   string `json:"url"`
	State string `json:"state"`
}

// Statuses probes every configured MCP app for userID by listing its tools.
func (g *Gateway) Statuses(ctx context.Context, userID int64) []AppStatus {
	apps := g.cfg.EnabledMCP()
	out := make([]AppStatus, len(apps))
	var wg sync.WaitGroup
	for i, app := range apps {
		wg.Go(func() {
			out[i] = AppStatus{App: app, Name: Label(app), URL: g.cfg.MCP[app].Site(), State: stateOf(g.mcp.Probe(ctx, userID, app))}
		})
	}
	wg.Wait()
	return out
}

func stateOf(err error) string {
	switch {
	case err == nil:
		return StateConnected
	case errors.Is(err, ErrNoAccount):
		return StateNoAccount
	case errors.Is(err, ErrNotConnected):
		return StateReconnect
	default:
		return StateUnavailable
	}
}
