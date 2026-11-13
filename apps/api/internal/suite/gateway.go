// Package suite connects Bulle to the Facile suite apps: it mints delegated
// tokens for each user and exposes the apps as agent tools.
package suite

type Gateway struct {
	tokens *Tokens
	cfg    Config
	mcp    *MCP
	spaces flightGroup
}

func NewGateway(tokens *Tokens, cfg Config) *Gateway {
	return &Gateway{tokens: tokens, cfg: cfg, mcp: NewMCP(cfg, tokens)}
}

func (g *Gateway) Config() Config { return g.cfg }

// MCP is the gateway's client of the suite MCP servers.
func (g *Gateway) MCP() *MCP { return g.mcp }

func (g *Gateway) Tokens() *Tokens { return g.tokens }
