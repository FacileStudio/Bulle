package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"

	"github.com/FacileStudio/Bulle/apps/api/internal/credentials"
	"github.com/FacileStudio/Bulle/core/provider"
	"github.com/FacileStudio/Bulle/core/sandbox"
	"github.com/FacileStudio/nacelle"
	nacelleanthropic "github.com/FacileStudio/nacelle/anthropic"
	"github.com/FacileStudio/nacelle/tools"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type ComputerProvider interface {
	Ensure(ctx context.Context, userID string) (*sandbox.Target, error)
	Touch(userID string)
}

type ToolBuilder func(target *sandbox.Target) ([]nacelle.Tool, io.Closer, error)

type BackendFactory func(backend, model string) (nacelle.Backend, error)

type CleanRunner struct{}

func (CleanRunner) env() []string {
	return []string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME")}
}

func (c CleanRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = c.env()
	return cmd.CombinedOutput()
}

func (c CleanRunner) RunInteractive(_ context.Context, cmd *exec.Cmd) error {
	cmd.Env = c.env()
	return cmd.Run()
}

func RemoteVMTools(target *sandbox.Target) ([]nacelle.Tool, io.Closer, error) {
	if target == nil {
		return nil, nil, fmt.Errorf("runner: no computer target")
	}
	return sandbox.RemoteTools(sandbox.ToolsOptions{Target: target, WorkDir: target.Workdir, Runner: CleanRunner{}})
}

func WebFetchTools(int64) ([]nacelle.Tool, error) {
	return tools.WebFetch()
}

type noComputer struct{}

func (noComputer) Ensure(context.Context, string) (*sandbox.Target, error) {
	return nil, fmt.Errorf("no computer provider is configured")
}

func (noComputer) Touch(string) {}

func (r *Runner) backend(ctx context.Context, userID int64, backend, model string) (nacelle.Backend, error) {
	if r.cfg.Backends != nil {
		return r.cfg.Backends(backend, model)
	}
	if !slices.Contains(credentials.Backends(), backend) {
		return nil, fmt.Errorf("unknown backend %q", backend)
	}
	key, err := r.cfg.Credentials.Resolve(ctx, userID, backend)
	if err != nil {
		if errors.Is(err, credentials.ErrNoKey) || errors.Is(err, credentials.ErrUnknownBackend) {
			return nil, fmt.Errorf("backend %s has no key configured", backend)
		}
		return nil, err
	}
	return newBackend(backend, model, key)
}

func EnvBackends(backend, model string) (nacelle.Backend, error) {
	if !slices.Contains(credentials.Backends(), backend) {
		return nil, fmt.Errorf("unknown backend %q", backend)
	}
	key := credentials.EnvKey(backend)
	if key == "" {
		return nil, fmt.Errorf("backend %s has no server key configured", backend)
	}
	return newBackend(backend, model, key)
}

func newBackend(backend, model, key string) (nacelle.Backend, error) {
	var baseURL string
	if backend == "openrouter" {
		baseURL = os.Getenv("OPENROUTER_BASE_URL")
	} else if backend == "openai" {
		baseURL = os.Getenv("OPENAI_BASE_URL")
	}
	if backend == "anthropic" {
		return nacelleanthropic.New(nacelleanthropic.Config{
			Model:   model,
			Options: []option.RequestOption{option.WithAPIKey(key)},
		}), nil
	}
	return provider.New(provider.Config{Backend: backend, Model: model, BaseURL: baseURL, APIKey: key})
}

func EnvSecrets() Redactor {
	var secrets []string
	for _, backend := range append(credentials.Backends(), credentials.SearchEngines()...) {
		for _, name := range credentials.EnvVars(backend) {
			secrets = append(secrets, os.Getenv(name))
		}
	}
	return NewRedactor(secrets...)
}

func (r *Runner) deviceTools(userID int64) ([]nacelle.Tool, string) {
	if r.cfg.Devices == nil {
		return nil, ""
	}
	return r.cfg.Devices(userID)
}
