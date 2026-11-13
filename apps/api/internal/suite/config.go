package suite

import "strings"

const (
	AppCourrier = "courrier"
	AppAgenda   = "agenda"
	AppOpus     = "opus"
	AppSablier  = "sablier"
	AppMycelium = "mycelium"
	AppNuage    = "nuage"
	AppGlouton  = "glouton"
	AppSonar    = "sonar"
	AppPlume    = "plume"
	AppCapsule  = "capsule"
	AppJournal  = "journal"
	AppAntenne  = "antenne"
	AppVision   = "vision"
	AppEcho     = "echo"
	AppToile    = "toile"
	AppDocs     = "docs"
	AppFeuilles = "feuilles"
)

// mcpNames are the apps Bulle reaches as an MCP client. Mycelium's tools stay
// behind Bulle's own memory and artifact tools (facadeMCP). Casier is not
// here, on purpose: an assistant never handles the owner's secrets.
var mcpNames = []string{AppCourrier, AppAgenda, AppSonar, AppNuage, AppOpus, AppSablier, AppGlouton, AppMycelium,
	AppPlume, AppCapsule, AppJournal, AppAntenne, AppVision, AppEcho, AppToile, AppDocs, AppFeuilles}

// facadeMCP are the MCP apps Bulle calls itself instead of mounting their
// tools for the model: Mycelium's memory and artifacts are scoped to the
// conversation's space and drawn as cards by Bulle's own tools.
var facadeMCP = map[string]bool{AppMycelium: true}

// MyceliumEnabled reports whether Mycelium's MCP server is configured.
func (c Config) MyceliumEnabled() bool { return c.MCP[AppMycelium].Enabled() }

type Config struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	MCP          map[string]MCPConfig
}

func FromEnv(issuer, clientID, clientSecret string) (Config, error) {
	cfg := Config{Issuer: issuer, ClientID: clientID, ClientSecret: clientSecret, MCP: map[string]MCPConfig{}}
	for _, name := range mcpNames {
		server, err := mcpFromEnv(name)
		if err != nil {
			return Config{}, err
		}
		cfg.MCP[name] = server
	}
	return cfg, nil
}

func Scope(app string) string { return app + ".read" }

func Label(app string) string {
	return strings.ToUpper(app[:1]) + app[1:]
}
