package suite

import (
	"fmt"
	"os"
	"strings"
)

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
)

// appNames are the apps Bulle still reaches through its built-in REST
// clients. mcpNames are the apps it reaches as an MCP client instead: an app
// moves from one list to the other when its MCP server replaces its client.
// Sablier is in both: its tools come from MCP, and the REST client only books
// the agent's own time, which no MCP tool covers. Mycelium is MCP only, but
// its tools stay behind Bulle's own memory and artifact tools (facadeMCP).
// Casier is in neither, on purpose: an assistant never handles the owner's
// secrets.
var (
	appNames = []string{AppSablier}
	mcpNames = []string{AppCourrier, AppAgenda, AppSonar, AppNuage, AppOpus, AppSablier, AppGlouton, AppMycelium,
		AppPlume, AppCapsule, AppJournal, AppAntenne, AppVision, AppEcho}
)

// facadeMCP are the MCP apps Bulle calls itself instead of mounting their
// tools for the model: Mycelium's memory and artifacts are scoped to the
// conversation's space and drawn as cards by Bulle's own tools.
var facadeMCP = map[string]bool{AppMycelium: true}

// MyceliumEnabled reports whether Mycelium's MCP server is configured.
func (c Config) MyceliumEnabled() bool { return c.MCP[AppMycelium].Enabled() }

type AppConfig struct {
	URL      string
	Audience string
}

func (a AppConfig) Enabled() bool { return a.URL != "" && a.Audience != "" }

type Config struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	Apps         map[string]AppConfig
	MCP          map[string]MCPConfig
}

func FromEnv(issuer, clientID, clientSecret string) (Config, error) {
	cfg := Config{Issuer: issuer, ClientID: clientID, ClientSecret: clientSecret, Apps: map[string]AppConfig{}, MCP: map[string]MCPConfig{}}
	for _, name := range appNames {
		upper := strings.ToUpper(name)
		app := AppConfig{
			URL:      strings.TrimRight(os.Getenv("SUITE_"+upper+"_URL"), "/"),
			Audience: os.Getenv("SUITE_" + upper + "_AUDIENCE"),
		}
		if (app.URL == "") != (app.Audience == "") {
			return Config{}, fmt.Errorf("SUITE_%s_URL and SUITE_%s_AUDIENCE must be set together", upper, upper)
		}
		cfg.Apps[name] = app
	}
	for _, name := range mcpNames {
		server, err := mcpFromEnv(name)
		if err != nil {
			return Config{}, err
		}
		cfg.MCP[name] = server
	}
	return cfg, nil
}

func (c Config) Enabled() []string {
	var names []string
	for _, name := range appNames {
		if c.Apps[name].Enabled() {
			names = append(names, name)
		}
	}
	return names
}

func Scope(app string) string { return app + ".read" }

func WriteScope(app string) string { return app + ".read " + app + ".write" }

func Label(app string) string {
	return strings.ToUpper(app[:1]) + app[1:]
}
