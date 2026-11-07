package suite

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

var gloutonWireKeys = map[string]string{
	"company_id": "companyId", "person_id": "personId", "first_name": "firstName", "last_name": "lastName",
}

// gloutonBody is a request body in Glouton's camelCase from the snake_case
// fields the tools advertise.
func gloutonBody(value any) (map[string]any, error) {
	body, err := toMap(value)
	if err != nil {
		return nil, err
	}
	renamed, _ := renameKeys(body, gloutonWireKeys).(map[string]any)
	return renamed, nil
}

func (g Glouton) read(ctx context.Context, userID int64, path string, query url.Values) ([]byte, error) {
	raw, err := g.gw.Get(ctx, userID, AppGlouton, path, query)
	if err != nil {
		return nil, err
	}
	return unwrapData(raw), nil
}

func (g Glouton) patch(ctx context.Context, userID int64, path string, body map[string]any) ([]byte, error) {
	if len(body) == 0 {
		return nil, errors.New("pass at least one field to change")
	}
	return g.write(ctx, userID, http.MethodPatch, path, body)
}

func (g Glouton) write(ctx context.Context, userID int64, method, path string, body any) ([]byte, error) {
	raw, err := g.gw.Do(ctx, userID, Request{App: AppGlouton, Method: method, Path: path, Body: body})
	if err != nil {
		return nil, err
	}
	return unwrapData(raw), nil
}

func gloutonItem(base, id string) (string, error) {
	if strings.TrimSpace(id) == "" {
		return "", errors.New("pass the id")
	}
	return base + "/" + url.PathEscape(id), nil
}

func gloutonLimit(limit int) int {
	if limit <= 0 {
		return 25
	}
	return min(limit, gloutonMaxLimit)
}
