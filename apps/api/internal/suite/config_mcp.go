package suite

import (
	"cmp"
	"fmt"
	"net/url"
	"os"
	"strings"
)

// MCPConfig is one suite app's MCP server. URL is where Bulle dials it and
// Resource the RFC 8707 resource the delegated token is minted for, which is
// the URL Registre declares for that server. They differ only when Bulle
// reaches the app on an internal address.
type MCPConfig struct {
	URL      string
	Resource string
}

// readOnlyMCP are the MCP servers that declare no write scope in Registre.
var readOnlyMCP = map[string]bool{AppSonar: true}

func (m MCPConfig) Enabled() bool { return m.URL != "" && m.Resource != "" }

// Site is the app's public origin, read from its resource, for the links Bulle
// shows the user.
func (m MCPConfig) Site() string {
	parsed, err := url.Parse(m.Resource)
	if err != nil {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host
}

// mcpFromEnv reads SUITE_<APP>_MCP_URL and SUITE_<APP>_MCP_RESOURCE. The
// resource defaults to the URL, which is right whenever Bulle dials the
// server at its public address.
func mcpFromEnv(name string) (MCPConfig, error) {
	upper := strings.ToUpper(name)
	server := MCPConfig{URL: strings.TrimSpace(os.Getenv("SUITE_" + upper + "_MCP_URL"))}
	server.Resource = cmp.Or(strings.TrimSpace(os.Getenv("SUITE_"+upper+"_MCP_RESOURCE")), server.URL)
	if server.URL == "" {
		return MCPConfig{}, nil
	}
	for _, value := range []string{server.URL, server.Resource} {
		parsed, err := url.Parse(value)
		if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
			return MCPConfig{}, fmt.Errorf("SUITE_%s_MCP_URL and SUITE_%s_MCP_RESOURCE must be absolute http(s) URLs", upper, upper)
		}
	}
	return server, nil
}

// EnabledMCP names the apps whose MCP server is configured.
func (c Config) EnabledMCP() []string {
	var names []string
	for _, name := range mcpNames {
		if c.MCP[name].Enabled() {
			names = append(names, name)
		}
	}
	return names
}
