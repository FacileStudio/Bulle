package suite

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/FacileStudio/porte"
)

const (
	exchangeGrant     = "urn:ietf:params:oauth:grant-type:token-exchange"
	accessTokenType   = "urn:ietf:params:oauth:token-type:access_token"
	maxRegistreBody   = 1 << 20
	registreCallLimit = 15 * time.Second
)

type Delegated struct {
	Token  string
	Expiry time.Time
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Error        string `json:"error"`
}

type Registre struct {
	cfg    Config
	client *http.Client
	now    func() time.Time

	mu       sync.Mutex
	endpoint string
}

func newRegistre(cfg Config, client *http.Client, now func() time.Time) *Registre {
	if client == nil {
		client = &http.Client{Timeout: registreCallLimit}
	}
	return &Registre{cfg: cfg, client: client, now: now}
}

func (r *Registre) tokenEndpoint(ctx context.Context) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.endpoint != "" {
		return r.endpoint, nil
	}
	target := strings.TrimRight(r.cfg.Issuer, "/") + "/.well-known/openid-configuration"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return "", err
	}
	response, err := r.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("suite: registre discovery: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	var document struct {
		TokenEndpoint string `json:"token_endpoint"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, maxRegistreBody)).Decode(&document); err != nil || document.TokenEndpoint == "" {
		return "", fmt.Errorf("suite: registre discovery returned no token endpoint")
	}
	r.endpoint = document.TokenEndpoint
	return r.endpoint, nil
}

func (r *Registre) post(ctx context.Context, form url.Values) (tokenResponse, error) {
	endpoint, err := r.tokenEndpoint(ctx)
	if err != nil {
		return tokenResponse{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return tokenResponse{}, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(url.QueryEscape(r.cfg.ClientID), url.QueryEscape(r.cfg.ClientSecret))
	response, err := r.client.Do(request)
	if err != nil {
		return tokenResponse{}, fmt.Errorf("suite: registre token call: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	var out tokenResponse
	_ = json.NewDecoder(io.LimitReader(response.Body, maxRegistreBody)).Decode(&out)
	if response.StatusCode >= 500 {
		return tokenResponse{}, fmt.Errorf("suite: registre answered %d", response.StatusCode)
	}
	if response.StatusCode != http.StatusOK || out.AccessToken == "" {
		return out, fmt.Errorf("suite: registre refused the request: %s", cleanOAuthError(out.Error))
	}
	return out, nil
}

func cleanOAuthError(code string) string {
	if code == "" {
		return "unknown"
	}
	return code
}

func (r *Registre) Refresh(ctx context.Context, refreshToken string) (porte.TokenSet, error) {
	out, err := r.post(ctx, url.Values{"grant_type": {"refresh_token"}, "refresh_token": {refreshToken}})
	if err != nil {
		if out.Error == "invalid_grant" {
			return porte.TokenSet{}, ErrNotConnected
		}
		return porte.TokenSet{}, err
	}
	next := out.RefreshToken
	if next == "" {
		next = refreshToken
	}
	return porte.TokenSet{
		AccessToken:  out.AccessToken,
		RefreshToken: next,
		Expiry:       r.now().Add(time.Duration(out.ExpiresIn) * time.Second),
	}, nil
}

func (r *Registre) Exchange(ctx context.Context, subjectToken, app string) (Delegated, error) {
	form := url.Values{
		"grant_type":         {exchangeGrant},
		"subject_token":      {subjectToken},
		"subject_token_type": {accessTokenType},
		"audience":           {r.cfg.Apps[app].Audience},
		"scope":              {Scope(app)},
	}
	out, err := r.post(ctx, form)
	if err != nil {
		if out.Error == "invalid_grant" {
			return Delegated{}, ErrSubjectRejected
		}
		return Delegated{}, err
	}
	return Delegated{Token: out.AccessToken, Expiry: r.now().Add(time.Duration(out.ExpiresIn) * time.Second)}, nil
}
