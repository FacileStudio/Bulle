package suite

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	maxAppBody   = 4 << 20
	appCallLimit = 20 * time.Second
)

type Gateway struct {
	tokens *Tokens
	cfg    Config
	client *http.Client
	spaces flightGroup
}

func NewGateway(tokens *Tokens, cfg Config, client *http.Client) *Gateway {
	if client == nil {
		client = &http.Client{Timeout: appCallLimit}
	}
	return &Gateway{tokens: tokens, cfg: cfg, client: client}
}

func (g *Gateway) Config() Config { return g.cfg }

func (g *Gateway) Tokens() *Tokens { return g.tokens }

func (g *Gateway) Get(ctx context.Context, userID int64, app, path string, query url.Values) ([]byte, error) {
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return g.Do(ctx, userID, Request{App: app, Method: http.MethodGet, Path: path})
}

func (g *Gateway) Do(ctx context.Context, userID int64, req Request) ([]byte, error) {
	app := req.App
	body, contentType, err := req.encode()
	if err != nil {
		return nil, err
	}
	call := apiCall{user: userID, app: app, method: req.Method, target: g.cfg.Apps[app].URL + req.Path, body: body, contentType: contentType}
	raw, status, err := g.call(ctx, call)
	if status == http.StatusUnauthorized {
		g.tokens.Invalidate(userID, app)
		raw, status, err = g.call(ctx, call)
	}
	return g.outcome(app, raw, status, err)
}

type apiCall struct {
	user        int64
	app         string
	method      string
	target      string
	body        []byte
	contentType string
}

func (g *Gateway) token(ctx context.Context, c apiCall) (string, error) {
	if c.method == http.MethodGet {
		return g.tokens.For(ctx, c.user, c.app)
	}
	return g.tokens.ForWrite(ctx, c.user, c.app)
}

func (g *Gateway) call(ctx context.Context, c apiCall) ([]byte, int, error) {
	token, err := g.token(ctx, c)
	if err != nil {
		return nil, 0, err
	}
	var reader io.Reader
	if c.body != nil {
		reader = bytes.NewReader(c.body)
	}
	request, err := http.NewRequestWithContext(ctx, c.method, c.target, reader)
	if err != nil {
		return nil, 0, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Accept", "application/json")
	if c.contentType != "" {
		request.Header.Set("Content-Type", c.contentType)
	}
	response, err := g.client.Do(request)
	if err != nil {
		return nil, 0, &AppError{App: c.app, URL: g.cfg.Apps[c.app].URL, Kind: ErrAppDown}
	}
	defer func() { _ = response.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(response.Body, maxAppBody))
	return raw, response.StatusCode, err
}

func (g *Gateway) outcome(app string, body []byte, status int, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	site := g.cfg.Apps[app].URL
	switch {
	case status == http.StatusUnauthorized:
		return nil, &AppError{App: app, URL: site, Kind: ErrNoAccount}
	case status >= http.StatusInternalServerError:
		return nil, &AppError{App: app, URL: site, Kind: ErrAppDown}
	case status < http.StatusOK || status >= http.StatusMultipleChoices:
		return nil, &RefusedError{App: app, Status: status, Reason: refusal(body)}
	}
	return body, nil
}
