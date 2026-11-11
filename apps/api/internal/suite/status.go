package suite

import (
	"context"
	"errors"
	"slices"
	"sync"
)

const (
	StateConnected   = "connected"
	StateNoAccount   = "no_account"
	StateReconnect   = "reconnect"
	StateUnavailable = "unavailable"
)

var probePaths = map[string]string{
	AppMycelium: myceliumSpacesPath,
}

type AppStatus struct {
	App   string `json:"app"`
	Name  string `json:"name"`
	URL   string `json:"url"`
	State string `json:"state"`
}

// Statuses probes every configured app for userID: an MCP app by listing its
// tools, any other by one read of its API. A REST client with no probe, such
// as Sablier's agent-time client, is covered by its MCP server's state.
func (g *Gateway) Statuses(ctx context.Context, userID int64) []AppStatus {
	mcpApps := g.cfg.EnabledMCP()
	apps := slices.Clone(mcpApps)
	for _, app := range g.cfg.Enabled() {
		if probePaths[app] != "" {
			apps = append(apps, app)
		}
	}
	out := make([]AppStatus, len(apps))
	var wg sync.WaitGroup
	for i, app := range apps {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if i < len(mcpApps) {
				out[i] = AppStatus{App: app, Name: Label(app), URL: g.cfg.MCP[app].Site(), State: stateOf(g.mcp.Probe(ctx, userID, app))}
				return
			}
			_, err := g.Get(ctx, userID, app, probePaths[app], nil)
			out[i] = AppStatus{App: app, Name: Label(app), URL: g.cfg.Apps[app].URL, State: stateOf(err)}
		}()
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
