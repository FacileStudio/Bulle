package env

import (
	"cmp"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/porte"
)

type Config struct {
	DatabaseURL        string
	Port               string
	CORSAllowedOrigins []string
	LogLevel           string
	JournalURL         string
	JournalToken       string
	Porte              porte.Config
	Suite              suite.Config
	Boite              Boite
	LLM                LLM
}

type LLM struct {
	Backend string
	Model   string
}

type Boite struct {
	WorkerSock  string
	WorkerToken string
	BoxKey      string
	IdleMinutes int
	BoxCPUs     int
	BoxMemoryMB int
	BoxDiskGB   int
}

func Load() (Config, error) {
	cfg := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        valueOrDefault("PORT", "4000"),
		LogLevel:    valueOrDefault("LOG_LEVEL", "info"),
		CORSAllowedOrigins: csvOrDefault("CORS_ALLOWED_ORIGINS", []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
		}),
		JournalURL:   os.Getenv("JOURNAL_URL"),
		JournalToken: os.Getenv("JOURNAL_TOKEN"),
		Porte: porte.Config{
			Issuer:       os.Getenv("OIDC_ISSUER"),
			ClientID:     os.Getenv("OIDC_CLIENT_ID"),
			ClientSecret: os.Getenv("OIDC_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("OIDC_REDIRECT_URL"),
			SuccessURL:   os.Getenv("OIDC_SUCCESS_URL"),
			CLIAudience:  os.Getenv("OIDC_CLI_AUDIENCE"),
		},
	}

	suiteConfig, err := suite.FromEnv(cfg.Porte.Issuer, cfg.Porte.ClientID, cfg.Porte.ClientSecret)
	if err != nil {
		return Config{}, err
	}
	cfg.Suite = suiteConfig

	boite, err := loadBoite()
	if err != nil {
		return Config{}, err
	}
	cfg.Boite = boite

	cfg.LLM = LLM{
		Backend: valueOrDefault("BULLE_LLM_BACKEND", valueOrDefault("LLM_BACKEND", "openrouter")),
		Model:   valueOrDefault("BULLE_LLM_MODEL", valueOrDefault("LLM_MODEL", "anthropic/claude-haiku-5.5")),
	}

	ssoOnly, err := boolOrDefault("SSO_ONLY", true)
	if err != nil {
		return Config{}, err
	}
	cfg.Porte.SSOOnly = ssoOnly

	if !cfg.Porte.Enabled() {
		return Config{}, fmt.Errorf("OIDC_ISSUER is required: Bulle signs in through Registre only")
	}
	if err := cfg.Porte.Validate(); err != nil {
		return Config{}, err
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	port, err := strconv.Atoi(cfg.Port)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("PORT must be a valid TCP port")
	}
	if err := validateOrigins(cfg.CORSAllowedOrigins); err != nil {
		return Config{}, err
	}
	if err := validateLogLevel(cfg.LogLevel); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func boolOrDefault(key string, fallback bool) (bool, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean", key)
	}
	return parsed, nil
}

func valueOrDefault(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func csvOrDefault(key string, fallback []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

func validateOrigins(origins []string) error {
	if len(origins) == 0 {
		return fmt.Errorf("CORS_ALLOWED_ORIGINS must contain at least one origin")
	}
	for _, origin := range origins {
		if origin == "*" || strings.HasPrefix(origin, "http://") || strings.HasPrefix(origin, "https://") {
			continue
		}
		return fmt.Errorf("CORS_ALLOWED_ORIGINS contains invalid origin %q", origin)
	}
	return nil
}

func validateLogLevel(level string) error {
	switch strings.ToLower(level) {
	case "debug", "info", "warn", "error":
		return nil
	default:
		return fmt.Errorf("LOG_LEVEL must be one of debug, info, warn, error")
	}
}

// loadBoite reads the box settings. The control key is read under its own name
// first and under the older BOITE_CONTROL_KEY second, because a deployment that
// still holds only the old secret would otherwise start without a key and fail
// every run at the box probe.
func loadBoite() (Boite, error) {
	boite := Boite{
		WorkerSock:  os.Getenv("BOITE_WORKER_SOCK"),
		WorkerToken: os.Getenv("BOITE_WORKER_TOKEN"),
		BoxKey:      cmp.Or(os.Getenv("BULLE_BOX_KEY"), os.Getenv("BOITE_CONTROL_KEY")),
	}
	for _, field := range []struct {
		key      string
		fallback int
		target   *int
	}{
		{"BULLE_IDLE_MINUTES", 30, &boite.IdleMinutes},
		{"BULLE_BOX_CPUS", 2, &boite.BoxCPUs},
		{"BULLE_BOX_MEMORY_MB", 4096, &boite.BoxMemoryMB},
		{"BULLE_BOX_DISK_GB", 20, &boite.BoxDiskGB},
	} {
		value, err := positiveIntOrDefault(field.key, field.fallback)
		if err != nil {
			return Boite{}, err
		}
		*field.target = value
	}
	return boite, nil
}

func positiveIntOrDefault(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return parsed, nil
}
