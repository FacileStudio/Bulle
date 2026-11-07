package env

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/FacileStudio/Bulle/core/jev"
)

// AI is what POST /api/ai/decide runs on: the decision model, whose key still
// comes from TYPESAFE_API_KEY (or OPENROUTER_API_KEY on the decisions
// endpoint), and the services allowed to call it, each named with the SHA-256
// of its key.
type AI struct {
	Jev     jev.Config
	Callers map[string][]byte
}

func loadAI() (AI, error) {
	cfg := jev.Config{
		Endpoint: jev.Endpoint(strings.TrimSpace(os.Getenv("JEV_ENDPOINT"))),
		BaseURL:  strings.TrimSpace(os.Getenv("JEV_BASE_URL")),
		Model:    strings.TrimSpace(os.Getenv("JEV_MODEL")),
	}
	switch cfg.Endpoint {
	case "", jev.EndpointSystemOne, jev.EndpointDecisions:
	default:
		return AI{}, fmt.Errorf("JEV_ENDPOINT must be systemone or decisions")
	}
	if cfg.BaseURL != "" {
		parsed, err := url.Parse(cfg.BaseURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return AI{}, fmt.Errorf("JEV_BASE_URL must be an http or https URL")
		}
	}
	callers, err := aiCallers(csvOrDefault("BULLE_AI_CALLERS", nil))
	return AI{Jev: cfg, Callers: callers}, err
}

func aiCallers(entries []string) (map[string][]byte, error) {
	callers := make(map[string][]byte, len(entries))
	for _, entry := range entries {
		name, digest, _ := strings.Cut(entry, ":")
		name = strings.TrimSpace(name)
		sum, err := hex.DecodeString(strings.TrimSpace(digest))
		if name == "" || err != nil || len(sum) != sha256.Size {
			return nil, fmt.Errorf("BULLE_AI_CALLERS entries must be name:<sha256 hex of the key>")
		}
		if _, taken := callers[name]; taken {
			return nil, fmt.Errorf("BULLE_AI_CALLERS names %q twice", name)
		}
		callers[name] = sum
	}
	return callers, nil
}
