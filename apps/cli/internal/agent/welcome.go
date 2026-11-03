package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/FacileStudio/blob"

	"github.com/FacileStudio/Bulle/core/settings"
)

const (
	welcomePixels  = 22
	welcomeTimeout = 1500 * time.Millisecond
	defaultName    = "Bulle"
)

// welcome is the start message. A signed-in session opens on the account's
// assistant drawn as its blob, the same one the web app shows, next to its
// name. A configured start_message, no login, NO_COLOR or a server that does
// not answer in time all keep the configured text.
func welcome(config settings.Config) string {
	configured := *config.StartMessage
	if configured != *settings.Defaults("").StartMessage || os.Getenv("NO_COLOR") != "" {
		return configured
	}
	name, ok := assistantName(config.Server.URL, settings.Token())
	if !ok {
		return configured
	}
	return blobBeside(name)
}

// assistantName asks the server who the stored token's assistant is.
func assistantName(rawServer, token string) (string, bool) {
	server, err := settings.NormalizeServerURL(rawServer)
	if err != nil || token == "" {
		return "", false
	}
	ctx, cancel := context.WithTimeout(context.Background(), welcomeTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server+"/api/auth/me", nil)
	if err != nil {
		return "", false
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", false
	}
	var me struct {
		User struct {
			Assistant *struct {
				Name string `json:"name"`
			} `json:"assistant"`
		} `json:"user"`
	}
	if json.NewDecoder(resp.Body).Decode(&me) != nil {
		return "", false
	}
	if me.User.Assistant == nil || strings.TrimSpace(me.User.Assistant.Name) == "" {
		return defaultName, true
	}
	return strings.TrimSpace(me.User.Assistant.Name), true
}

// blobBeside draws the blob a name seeds with the name on its middle row.
func blobBeside(name string) string {
	rows := strings.Split(blob.New(name, blob.Options{}).Terminal(welcomePixels), "\n")
	mid := len(rows) / 2
	rows[mid] += "  \x1b[1m" + name + "\x1b[0m"
	return strings.Join(rows, "\n")
}
