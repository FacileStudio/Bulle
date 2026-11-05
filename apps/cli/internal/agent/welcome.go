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
	welcomePixels  = 16
	welcomeTimeout = 1500 * time.Millisecond
	defaultName    = "Bulle"
)

// welcome is the start message. Unless start_message is set, the session opens
// on the account's assistant drawn as its blob, the same one the web app shows;
// signed out, it draws the default name. NO_COLOR keeps the configured text.
func welcome(config settings.Config, assistant string) string {
	configured := *config.StartMessage
	if configured != "" || os.Getenv("NO_COLOR") != "" {
		return configured
	}
	if assistant == "" {
		assistant = defaultName
	}
	return blob.New(assistant, blob.Options{}).Terminal(welcomePixels)
}

// signedInAssistant names the stored token's assistant, or "" when there is no
// login or the server does not answer in time.
func signedInAssistant(config settings.Config) string {
	name, _ := assistantName(config.Server.URL, settings.Token())
	return name
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
	defer func() { _ = resp.Body.Close() }()
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
