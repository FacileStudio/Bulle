package settings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"
)

// ErrNotSignedIn is the refusal a session gets with no server login: the
// server owns the provider, its keys and its models.
var ErrNotSignedIn = errors.New("bulle runs on a Bulle server: run `bulle login <url>` first")

const llmTimeout = 5 * time.Second

// LLM is the model setup a server offers.
type LLM struct {
	Backend string   `json:"backend"`
	Model   string   `json:"model"`
	Models  []string `json:"models"`
	Persona string   `json:"persona"`
}

// FetchLLM asks the server which backend and models it offers.
func FetchLLM(server, token string) (LLM, error) {
	ctx, cancel := context.WithTimeout(context.Background(), llmTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server+"/api/llm", nil)
	if err != nil {
		return LLM{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return LLM{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return LLM{}, fmt.Errorf("%s/api/llm answered %s", server, resp.Status)
	}
	var llm LLM
	if err := json.NewDecoder(resp.Body).Decode(&llm); err != nil {
		return LLM{}, fmt.Errorf("reading %s/api/llm: %w", server, err)
	}
	if llm.Backend == "" || llm.Model == "" {
		return LLM{}, fmt.Errorf("%s/api/llm names no backend or model", server)
	}
	return llm, nil
}

// UseServer points c at the server's proxy, on the backend and models the
// server offers. A model c already names (a job's, or a /model pick) is kept
// only when the server offers it.
func UseServer(c *Config) error {
	token := Token()
	if c.Server.URL == "" || token == "" {
		return ErrNotSignedIn
	}
	base, err := NormalizeServerURL(c.Server.URL)
	if err != nil {
		return err
	}
	llm, err := FetchLLM(base, token)
	if err != nil {
		return err
	}
	model := llm.Model
	if c.Model != "" {
		if !slices.Contains(llm.Models, c.Model) {
			return fmt.Errorf("model %q is not offered by %s; it offers %s", c.Model, base, strings.Join(llm.Models, ", "))
		}
		model = c.Model
	}
	c.Provider = Provider{Backend: llm.Backend, Model: model, BaseURL: base + "/api/llm/" + llm.Backend, APIKey: token, Models: llm.Models, Persona: llm.Persona}
	return nil
}
