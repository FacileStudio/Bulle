// Package provider constructs model backends and inspects available catalogs.
package provider

import (
	"fmt"
	"strings"

	"github.com/FacileStudio/nacelle"
	"github.com/FacileStudio/nacelle/anthropic"
	"github.com/FacileStudio/nacelle/google"
	"github.com/FacileStudio/nacelle/openai"
	"github.com/FacileStudio/nacelle/openrouter"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Config holds connection parameters for building a backend.
type Config struct {
	Backend string
	Model   string
	BaseURL string
	APIKey  string
}

// New constructs the requested nacelle.Backend.
func New(cfg Config) (nacelle.Backend, error) {
	switch cfg.Backend {
	case "anthropic":
		if cfg.BaseURL == "" && cfg.APIKey != "" {
			return nil, fmt.Errorf("anthropic reads its key from ANTHROPIC_API_KEY: api_key applies to anthropic only with a base_url")
		}
		return anthropic.New(anthropic.Config{Model: cfg.Model, Options: anthropicOptions(cfg)}), nil
	case "google":
		return google.New(google.Config{
			Model:   cfg.Model,
			BaseURL: cfg.BaseURL,
			APIKey:  cfg.APIKey,
		})
	case "openai":
		return openai.New(openai.Config{
			Model:   cfg.Model,
			BaseURL: cfg.BaseURL,
			APIKey:  cfg.APIKey,
		})
	case "openrouter":
		if cfg.Model == "" {
			return nil, fmt.Errorf("openrouter needs a model")
		}
		return openrouter.New(openrouter.Config{
			Model:   cfg.Model,
			BaseURL: cfg.BaseURL,
			APIKey:  cfg.APIKey,
		})
	default:
		return nil, fmt.Errorf("unknown backend %q, want anthropic, google, openai, or openrouter", cfg.Backend)
	}
}

func anthropicOptions(cfg Config) []option.RequestOption {
	if cfg.BaseURL == "" {
		return nil
	}
	options := []option.RequestOption{option.WithBaseURL(strings.TrimRight(cfg.BaseURL, "/") + "/")}
	if cfg.APIKey != "" {
		options = append(options, option.WithAuthToken(cfg.APIKey), option.WithHeaderDel("X-Api-Key"))
	}
	return options
}
