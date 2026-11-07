package runner

import (
	"log/slog"
	"time"
)

const (
	defaultMaxRunsPerUser = 2
	defaultRunTimeout     = 30 * time.Minute
	defaultMaxIterations  = 50
)

// withDefaults fills in every field New does not require: the providers a
// test or a minimal deployment leaves out, the run limits, and the tool loop's
// iteration cap, which is never unbounded.
func (cfg Config) withDefaults() Config {
	cfg = cfg.withProviders()
	if cfg.Backend == "" {
		cfg.Backend = "anthropic"
	}
	if cfg.System == "" {
		cfg.System = defaultSystem
	}
	if cfg.MaxRunsPerUser <= 0 {
		cfg.MaxRunsPerUser = defaultMaxRunsPerUser
	}
	if cfg.RunTimeout <= 0 {
		cfg.RunTimeout = defaultRunTimeout
	}
	if cfg.MaxIterations <= 0 {
		cfg.MaxIterations = defaultMaxIterations
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	return cfg
}

func (cfg Config) withProviders() Config {
	if cfg.Computers == nil {
		cfg.Computers = noComputer{}
	}
	if cfg.Backends == nil && cfg.Credentials == nil {
		cfg.Backends = EnvBackends
	}
	if cfg.VMTools == nil {
		cfg.VMTools = RemoteVMTools
	}
	if cfg.HostTools == nil {
		cfg.HostTools = WebFetchTools
	}
	return cfg
}
