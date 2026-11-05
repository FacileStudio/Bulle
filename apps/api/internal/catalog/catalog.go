// Package catalog lists the models a backend actually offers, so the web
// settings and the CLI answer from the same list instead of each keeping one.
package catalog

import (
	"cmp"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Model is one entry a client may choose.
type Model struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Resolver is the credential store the catalogue asks for the key to call the
// vendor with. A vendor whose list is public (OpenRouter) answers without one.
type Resolver interface {
	Resolve(ctx context.Context, userID int64, backend string) (string, error)
}

const (
	// ttl keeps a vendor's catalogue for ten minutes: long enough that a
	// settings page or a /model list costs nothing, short enough that a model
	// the vendor publishes is picked up the same day.
	ttl = 10 * time.Minute

	requestTimeout = 10 * time.Second
	maxBody        = 4 << 20
)

type entry struct {
	models []Model
	at     time.Time
}

// Catalogue caches each (user, backend) list so the settings page, the CLI and
// /api/llm share one fetch rather than one per request.
type Catalogue struct {
	resolver Resolver
	fallback []Model
	client   *http.Client
	now      func() time.Time

	mu    sync.Mutex
	cache map[string]entry
}

// New builds a catalogue that falls back to fallback — the operator's
// BULLE_LLM_MODELS — when the vendor cannot be asked.
func New(resolver Resolver, fallback []string) *Catalogue {
	models := make([]Model, 0, len(fallback))
	for _, id := range fallback {
		if id = strings.TrimSpace(id); id != "" {
			models = append(models, Model{ID: id, Name: id})
		}
	}
	return &Catalogue{
		resolver: resolver,
		fallback: models,
		client:   &http.Client{Timeout: requestTimeout},
		now:      time.Now,
		cache:    map[string]entry{},
	}
}

// List returns the backend's models, falling back to the operator's list when
// the vendor answers nothing usable. The result is never nil.
func (c *Catalogue) List(ctx context.Context, userID int64, backend string) []Model {
	key := strconv.FormatInt(userID, 10) + "/" + backend
	c.mu.Lock()
	cached, ok := c.cache[key]
	c.mu.Unlock()
	if ok && c.now().Sub(cached.at) < ttl {
		return cached.models
	}

	list := c.fetch(ctx, userID, backend)
	if len(list) == 0 {
		list = c.fallback
	}
	c.mu.Lock()
	c.cache[key] = entry{models: list, at: c.now()}
	c.mu.Unlock()
	return list
}

// IDs is List reduced to the ids a client may name.
func (c *Catalogue) IDs(ctx context.Context, userID int64, backend string) []string {
	list := c.List(ctx, userID, backend)
	out := make([]string, 0, len(list))
	for _, model := range list {
		out = append(out, model.ID)
	}
	return out
}

func (c *Catalogue) fetch(ctx context.Context, userID int64, backend string) []Model {
	vendor, ok := vendors[backend]
	if !ok {
		return nil
	}
	key := ""
	if c.resolver != nil {
		if resolved, err := c.resolver.Resolve(ctx, userID, backend); err == nil {
			key = resolved
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, vendor.url, nil)
	if err != nil {
		return nil
	}
	vendor.auth(req, key)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return nil
	}
	list, err := vendor.parse(raw)
	if err != nil {
		return nil
	}
	slices.SortFunc(list, func(a, b Model) int { return cmp.Compare(a.ID, b.ID) })
	return list
}

type vendor struct {
	url   string
	auth  func(*http.Request, string)
	parse func([]byte) ([]Model, error)
}

var vendors = map[string]vendor{
	"openrouter": {
		url:  "https://openrouter.ai/api/v1/models",
		auth: bearer,
		parse: func(raw []byte) ([]Model, error) {
			var body struct {
				Data []struct{ ID, Name string } `json:"data"`
			}
			if err := json.Unmarshal(raw, &body); err != nil {
				return nil, err
			}
			return named(body.Data, func(i int) (string, string) {
				return body.Data[i].ID, body.Data[i].Name
			}), nil
		},
	},
	"openai": {
		url:  "https://api.openai.com/v1/models",
		auth: bearer,
		parse: func(raw []byte) ([]Model, error) {
			var body struct {
				Data []struct{ ID string } `json:"data"`
			}
			if err := json.Unmarshal(raw, &body); err != nil {
				return nil, err
			}
			return named(body.Data, func(i int) (string, string) {
				return body.Data[i].ID, body.Data[i].ID
			}), nil
		},
	},
	"anthropic": {
		url: "https://api.anthropic.com/v1/models?limit=1000",
		auth: func(r *http.Request, key string) {
			r.Header.Set("x-api-key", key)
			r.Header.Set("anthropic-version", "2023-06-01")
		},
		parse: func(raw []byte) ([]Model, error) {
			var body struct {
				Data []struct {
					ID          string `json:"id"`
					DisplayName string `json:"display_name"`
				} `json:"data"`
			}
			if err := json.Unmarshal(raw, &body); err != nil {
				return nil, err
			}
			return named(body.Data, func(i int) (string, string) {
				return body.Data[i].ID, body.Data[i].DisplayName
			}), nil
		},
	},
	"google": {
		url: "https://generativelanguage.googleapis.com/v1beta/models?pageSize=1000",
		auth: func(r *http.Request, key string) {
			r.Header.Set("x-goog-api-key", key)
		},
		parse: func(raw []byte) ([]Model, error) {
			var body struct {
				Models []struct {
					Name                       string   `json:"name"`
					DisplayName                string   `json:"displayName"`
					SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
				} `json:"models"`
			}
			if err := json.Unmarshal(raw, &body); err != nil {
				return nil, err
			}
			out := make([]Model, 0, len(body.Models))
			for _, m := range body.Models {
				if slices.Contains(m.SupportedGenerationMethods, "embedContent") &&
					!slices.Contains(m.SupportedGenerationMethods, "generateContent") {
					continue
				}
				id := strings.TrimPrefix(m.Name, "models/")
				out = append(out, Model{ID: id, Name: cmp.Or(m.DisplayName, id)})
			}
			return out, nil
		},
	},
}

func bearer(r *http.Request, key string) {
	if key != "" {
		r.Header.Set("Authorization", "Bearer "+key)
	}
}

// named turns a vendor's rows into models, dropping an entry with no id.
func named[T any](rows []T, pick func(int) (id, name string)) []Model {
	out := make([]Model, 0, len(rows))
	for i := range rows {
		id, name := pick(i)
		if id = strings.TrimSpace(id); id == "" {
			continue
		}
		out = append(out, Model{ID: id, Name: cmp.Or(strings.TrimSpace(name), id)})
	}
	return out
}
