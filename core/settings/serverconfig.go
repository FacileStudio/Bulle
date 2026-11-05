package settings

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// remoteTimeout bounds the one request a start makes for the CLI settings saved
// in the web app. It is short on purpose: a machine that cannot reach its
// server loses this layer, not its startup.
const remoteTimeout = 3 * time.Second

const maxRemoteBytes = 128 << 10

// RemoteConfig reads the CLI settings the signed-in user saved in the web
// settings (GET /api/cli/config) and returns them as the layer between the
// shipped defaults and ~/.config/bulle/config.yml. A local file still wins
// field by field, so one machine's exception survives what the browser saved,
// while a machine with no opinion picks the browser's answer up.
//
// Every failure returns an empty layer: no login, an offline server, a server
// older than this endpoint. A command that does not need the server must still
// run.
func RemoteConfig(serverURL, token string) Config {
	if serverURL == "" || token == "" {
		return Config{}
	}
	base, err := NormalizeServerURL(serverURL)
	if err != nil {
		return Config{}
	}
	ctx, cancel := context.WithTimeout(context.Background(), remoteTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/api/cli/config", nil)
	if err != nil {
		return Config{}
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Config{}
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return Config{}
	}
	var body struct {
		Config string `json:"config"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxRemoteBytes)).Decode(&body); err != nil {
		return Config{}
	}
	cfg, err := Parse([]byte(body.Config))
	if err != nil {
		return Config{}
	}
	// The address is where this layer was fetched from, never something the
	// layer may change: UseServer sends the login token to c.Server.URL.
	cfg.Server.URL = ""
	return cfg
}
