package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/FacileStudio/tronc/errors"
)

type Client struct {
	http  *http.Client
	token string
}

func New(socket, token string) *Client {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var dialer net.Dialer
			return dialer.DialContext(ctx, "unix", socket)
		},
	}
	return &Client{http: &http.Client{Transport: transport, Timeout: 30 * time.Second}, token: token}
}

func (c *Client) Create(ctx context.Context, request CreateRequest) (Box, error) {
	var box Box
	return box, c.do(ctx, http.MethodPost, "/boxes", request, &box)
}

func (c *Client) List(ctx context.Context, owner string) ([]Box, error) {
	path := "/boxes"
	if owner != "" {
		path += "?owner=" + url.QueryEscape(owner)
	}
	var boxes []Box
	return boxes, c.do(ctx, http.MethodGet, path, nil, &boxes)
}

func (c *Client) Get(ctx context.Context, id string) (Box, error) {
	var box Box
	return box, c.do(ctx, http.MethodGet, boxPath(id, ""), nil, &box)
}

func (c *Client) Start(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodPost, boxPath(id, "/start"), nil, nil)
}

func (c *Client) Stop(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodPost, boxPath(id, "/stop"), nil, nil)
}

func (c *Client) Remove(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodPost, boxPath(id, "/rm"), nil, nil)
}

func (c *Client) SetSSHKey(ctx context.Context, id, key string) error {
	return c.do(ctx, http.MethodPut, boxPath(id, "/ssh-key"), map[string]string{"ssh_key": key}, nil)
}

// SetNet sets a box's egress mode, open or offline. The box takes it on its
// next boot.
func (c *Client) SetNet(ctx context.Context, id, net string) error {
	return c.do(ctx, http.MethodPut, boxPath(id, "/net"), map[string]string{"net": net}, nil)
}

func (c *Client) Snapshot(ctx context.Context, id, tag string) error {
	return c.do(ctx, http.MethodPost, boxPath(id, "/snapshot"), map[string]string{"tag": tag}, nil)
}

func (c *Client) Rollback(ctx context.Context, id, tag string) error {
	return c.do(ctx, http.MethodPost, boxPath(id, "/rollback"), map[string]string{"tag": tag}, nil)
}

func (c *Client) Snapshots(ctx context.Context, id string) ([]Snapshot, error) {
	var snapshots []Snapshot
	return snapshots, c.do(ctx, http.MethodGet, boxPath(id, "/snapshots"), nil, &snapshots)
}

func boxPath(id, suffix string) string {
	return "/boxes/" + url.PathEscape(id) + suffix
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return errors.Internal("failed to encode the worker request", err)
		}
		reader = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, "http://boite-worker"+path, reader)
	if err != nil {
		return errors.Internal("failed to build the worker request", err)
	}
	request.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := c.http.Do(request)
	if err != nil {
		return unavailable(err)
	}
	defer response.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(response.Body, 4<<20))
	if err != nil {
		return unavailable(err)
	}
	if response.StatusCode >= 300 {
		failure := &Error{Status: response.StatusCode}
		if json.Unmarshal(payload, failure) != nil || failure.Code == "" {
			failure.Code = CodeInternal
			failure.Message = http.StatusText(response.StatusCode)
		}
		return envelope(failure)
	}
	if out == nil || len(bytes.TrimSpace(payload)) == 0 {
		return nil
	}
	if err := json.Unmarshal(payload, out); err != nil {
		return errors.Internal("the computer service returned an unreadable response", err)
	}
	return nil
}
