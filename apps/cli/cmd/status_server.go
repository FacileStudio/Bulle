package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/FacileStudio/Bulle/core/settings"
)

func checkServer(parent context.Context) []statusLine {
	cfg, err := settings.Settings("", settings.Config{})
	if err != nil || cfg.Server.URL == "" {
		return nil
	}
	server, err := settings.NormalizeServerURL(cfg.Server.URL)
	if err != nil {
		return []statusLine{{statusFail, "server", err.Error()}}
	}
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()
	if _, err := serverReachable(ctx, server); err != nil {
		return []statusLine{{statusFail, "server", server + " is unreachable: " + err.Error()}}
	}
	lines := []statusLine{{statusOK, "server", server + " reachable"}}
	token := settings.Token()
	if token == "" {
		return append(lines, statusLine{statusWarn, "token", "not signed in, run bulle login"})
	}
	user, code, err := fetchUser(ctx, server, token)
	switch {
	case err != nil:
		return append(lines, statusLine{statusFail, "token", err.Error()})
	case code != http.StatusOK:
		return append(lines, statusLine{statusFail, "token", fmt.Sprintf("rejected with status %d, run bulle login", code)})
	}
	lines = append(lines, statusLine{statusOK, "token", "valid for " + user.Email})
	return append(lines, credentialLine(ctx, server, token))
}

func credentialLine(ctx context.Context, server, token string) statusLine {
	rows, code, err := fetchCredentials(ctx, server, token)
	if err != nil || code != http.StatusOK {
		return statusLine{statusWarn, "keys", "server did not list provider keys"}
	}
	var held []string
	for _, row := range rows {
		if row.Source == "user" || row.Source == "instance" {
			held = append(held, row.Backend+" ("+row.Source+")")
		}
	}
	if len(held) == 0 {
		return statusLine{statusWarn, "keys", "no backend has a key on the server"}
	}
	return statusLine{statusOK, "keys", strings.Join(held, ", ")}
}
