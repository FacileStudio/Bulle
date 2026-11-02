package suite

import (
	"context"
	"fmt"
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
	target := g.cfg.Apps[app].URL + path
	if len(query) > 0 {
		target += "?" + query.Encode()
	}
	body, status, err := g.call(ctx, userID, app, target)
	if status == http.StatusUnauthorized {
		g.tokens.Invalidate(userID, app)
		body, status, err = g.call(ctx, userID, app, target)
	}
	return g.outcome(app, body, status, err)
}

func (g *Gateway) call(ctx context.Context, userID int64, app, target string) ([]byte, int, error) {
	token, err := g.tokens.For(ctx, userID, app)
	if err != nil {
		return nil, 0, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, 0, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Accept", "application/json")
	response, err := g.client.Do(request)
	if err != nil {
		return nil, 0, &AppError{App: app, URL: g.cfg.Apps[app].URL, Kind: ErrAppDown}
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(response.Body, maxAppBody))
	return body, response.StatusCode, err
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
		return nil, fmt.Errorf("%s answered %d", Label(app), status)
	}
	return body, nil
}
