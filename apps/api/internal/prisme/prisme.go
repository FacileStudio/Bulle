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

// Actor is the person a call acts for: the email Prisme keys accounts by
// and, when Bulle knows it, their Registre subject, which Prisme attaches to
// the account so a deletion notice finds it. An empty Subject sends none.
type Actor struct {
	Email   string
	Subject string
}

// CodeSubjectTaken is Prisme's 409 for a subject that belongs to another
// account. It heals on its own when the person changed their address in
// Registre: their next Prisme sign-in rewrites the email, and the next call
// succeeds. Every other refusal of the subject stands until someone acts.
const CodeSubjectTaken = "acting_subject_taken"

// Refusal is a JSON error Prisme answered a call with, in tronc's envelope.
// Nothing was written.
type Refusal struct {
	Call    string
	Status  int
	Code    string
	Message string
}

func (r *Refusal) Error() string {
	return fmt.Sprintf("prisme %s: %d %s: %s", r.Call, r.Status, r.Code, r.Message)
}

// Final reports a refusal a plain retry cannot change: a 400
// (acting_subject_invalid: the header is malformed or sent twice), a 409
// (acting_subject_mismatch: the account for this email holds another subject;
// acting_subject_taken: the subject belongs to another account) or a 412
// (acting_subject_unsupported: Prisme has no issuer). A 409 is never
// transient. A 500 is not final and is retried.
func (r *Refusal) Final() bool {
	switch r.Status {
	case http.StatusBadRequest, http.StatusConflict, http.StatusPreconditionFailed:
		return true
	}
	return false
}

// Account is a Matrix login for an agent.
type Account struct {
	Homeserver  string `json:"homeserver"`
	UserID      string `json:"user_id"`
	OwnerUserID string `json:"owner_user_id"`
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
func (c *Client) Agent(ctx context.Context, actor Actor, name, avatar string) (Account, error) {
	var out Account
	err := c.call(ctx, http.MethodPost, "/api/agents", actor, map[string]string{"name": name, "avatar_url": avatar}, &out)
	if err == nil && (out.UserID == "" || out.AccessToken == "") {
		err = fmt.Errorf("prisme: agent answer has no user id or token")
	}
	return out, err
}

// Spaces lists the user's spaces with their channels.
func (c *Client) Spaces(ctx context.Context, actor Actor) ([]Space, error) {
	var out []Space
	return out, c.call(ctx, http.MethodGet, "/api/spaces", actor, nil, &out)
}

// SetRooms joins the agent to exactly these rooms.
func (c *Client) SetRooms(ctx context.Context, actor Actor, agent string, rooms []string) error {
	path := "/api/agents/" + url.PathEscape(agent) + "/rooms"
	return c.call(ctx, http.MethodPut, path, actor, map[string][]string{"room_ids": rooms}, nil)
}

func (c *Client) call(ctx context.Context, method, path string, actor Actor, in, out any) error {
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
	req.Header.Set("X-Prisme-Acting-For", actor.Email)
	if actor.Subject != "" {
		req.Header.Set("X-Prisme-Acting-Subject", actor.Subject)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("prisme %s %s: %w", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode/100 != 2 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return refused(method+" "+path, resp, detail)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// refused reads a non-2xx answer: Prisme's JSON error becomes a *Refusal,
// anything else keeps the status line and the body.
func refused(call string, resp *http.Response, detail []byte) error {
	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(detail, &body) == nil && body.Error.Code != "" {
		return &Refusal{Call: call, Status: resp.StatusCode, Code: body.Error.Code, Message: body.Error.Message}
	}
	return fmt.Errorf("prisme %s: %s %s", call, resp.Status, strings.TrimSpace(string(detail)))
}
