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
)

var appNames = []string{AppCourrier, AppAgenda, AppOpus, AppSablier}

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
}

func FromEnv(issuer, clientID, clientSecret string) (Config, error) {
	cfg := Config{Issuer: issuer, ClientID: clientID, ClientSecret: clientSecret, Apps: map[string]AppConfig{}}
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

func Label(app string) string {
	return strings.ToUpper(app[:1]) + app[1:]
}
