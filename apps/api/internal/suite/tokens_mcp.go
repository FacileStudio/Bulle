package suite

import "context"

// ForMCP is the user's delegated token for app's MCP server: audience-bound
// to its resource, with the read and write scopes Registre grants Bulle there.
func (t *Tokens) ForMCP(ctx context.Context, userID int64, app string) (string, error) {
	return t.token(ctx, cacheKey{user: userID, app: app, write: true, mcp: true})
}

// InvalidateMCP drops the cached MCP token of one user and app, after the
// server refused it.
func (t *Tokens) InvalidateMCP(userID int64, app string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.cache, cacheKey{user: userID, app: app, write: true, mcp: true})
}

func (t *Tokens) configured(key cacheKey) bool {
	if key.mcp {
		return t.cfg.MCP[key.app].Enabled()
	}
	return t.cfg.Apps[key.app].Enabled()
}
