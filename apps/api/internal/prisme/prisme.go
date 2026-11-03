// Package prisme talks to Prisme, the suite's Matrix chat, as a trusted
// service acting for one user: it provisions the user's agent account and
// lists the rooms of their spaces, so nobody configures a bot by hand.
package prisme

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client calls one Prisme instance with its service key.
type Client struct {
	base string
	key  string
	http *http.Client
}

// Account is a Matrix login for an agent.
type Account struct {
	Homeserver  string `json:"homeserver"`
	UserID      string `json:"user_id"`
	AccessToken string `json:"access_token"`
	DeviceID    string `json:"device_id"`
}

// Channel is one room of a space.
type Channel struct {
	RoomID string `json:"room_id"`
	Name   string `json:"name"`
}

// Space is a Prisme space and its channels.
type Space struct {
	Name     string    `json:"name"`
	Channels []Channel `json:"channels"`
}

// New returns a client, or nil when Prisme is not configured.
func New(base, key string) *Client {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if base == "" || key == "" {
		return nil
	}
	return &Client{base: base, key: key, http: &http.Client{Timeout: 30 * time.Second}}
}

// Agent creates or returns the user's agent with this name.
func (c *Client) Agent(ctx context.Context, email, name, avatar string) (Account, error) {
	var out Account
	err := c.call(ctx, http.MethodPost, "/api/agents", email, map[string]string{"name": name, "avatar_url": avatar}, &out)
	if err == nil && (out.UserID == "" || out.AccessToken == "") {
		err = fmt.Errorf("prisme: agent answer has no user id or token")
	}
	return out, err
}

// Spaces lists the user's spaces with their channels.
func (c *Client) Spaces(ctx context.Context, email string) ([]Space, error) {
	var out []Space
	return out, c.call(ctx, http.MethodGet, "/api/spaces", email, nil, &out)
}

// SetRooms joins the agent to exactly these rooms.
func (c *Client) SetRooms(ctx context.Context, email, agent string, rooms []string) error {
	path := "/api/agents/" + url.PathEscape(agent) + "/rooms"
	return c.call(ctx, http.MethodPut, path, email, map[string][]string{"room_ids": rooms}, nil)
}

func (c *Client) call(ctx context.Context, method, path, email string, in, out any) error {
	var body io.Reader
	if in != nil {
		raw, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.key)
	req.Header.Set("X-Prisme-Acting-For", email)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("prisme %s %s: %w", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode/100 != 2 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("prisme %s %s: %s %s", method, path, resp.Status, strings.TrimSpace(string(detail)))
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
