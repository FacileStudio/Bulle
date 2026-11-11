package suite

import (
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/suite/myceliumtest"
)

// myceliumWorld is a gateway whose only app is the fake Mycelium MCP server.
func myceliumWorld(t *testing.T) (*Gateway, *myceliumtest.Server) {
	t.Helper()
	fake := myceliumtest.New(t)
	cfg := Config{Issuer: fake.Issuer(), ClientID: "bulle", ClientSecret: "s3cret", Apps: map[string]AppConfig{},
		MCP: map[string]MCPConfig{AppMycelium: {URL: fake.Resource(), Resource: fake.Resource()}}}
	gw := NewGateway(NewTokens(cfg, fake, nil), cfg, nil)
	t.Cleanup(func() { _ = gw.MCP().Close() })
	return gw, fake
}

// everyApp is the world's gateway with Mycelium configured too, for tests
// that only look at which tools register.
func everyApp(w *world) *Gateway {
	cfg := w.cfg
	cfg.MCP = map[string]MCPConfig{AppMycelium: {URL: w.app.URL + "/mcp", Resource: w.app.URL + "/mcp"}}
	return NewGateway(NewTokens(cfg, w.ids, nil), cfg, nil)
}
