package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const apiMount = "/api"

var serverClient = &http.Client{Timeout: 15 * time.Second}

type serverUser struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type credentialRow struct {
	Backend string `json:"backend"`
	Last4   string `json:"last4"`
	Source  string `json:"source"`
}

func serverCall(ctx context.Context, method, url, token string, body any, out any) (int, error) {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return 0, err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return 0, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := serverClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 || out == nil {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<16))
		return resp.StatusCode, nil
	}
	return resp.StatusCode, json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(out)
}

func exchangeCode(ctx context.Context, server, code string) (string, error) {
	var out struct {
		Token string `json:"token"`
	}
	status, err := serverCall(ctx, http.MethodPost, server+apiMount+"/auth/oidc/exchange", "", map[string]string{"code": code}, &out)
	if err != nil {
		return "", fmt.Errorf("exchanging the login code: %w", err)
	}
	if status != http.StatusOK || out.Token == "" {
		return "", fmt.Errorf("the server refused the login code (status %d)", status)
	}
	return out.Token, nil
}

func fetchUser(ctx context.Context, server, token string) (serverUser, int, error) {
	var out struct {
		User serverUser `json:"user"`
	}
	status, err := serverCall(ctx, http.MethodGet, server+apiMount+"/auth/me", token, nil, &out)
	return out.User, status, err
}

func fetchCredentials(ctx context.Context, server, token string) ([]credentialRow, int, error) {
	var rows []credentialRow
	status, err := serverCall(ctx, http.MethodGet, server+apiMount+"/credentials", token, nil, &rows)
	return rows, status, err
}

func serverReachable(ctx context.Context, server string) (int, error) {
	return serverCall(ctx, http.MethodGet, server+apiMount+"/auth/config", "", nil, nil)
}
