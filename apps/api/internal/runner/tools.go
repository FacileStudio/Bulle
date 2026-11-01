package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/FacileStudio/Bulle/core/provider"
	"github.com/FacileStudio/Bulle/core/sandbox"
	"github.com/FacileStudio/nacelle"
	"github.com/FacileStudio/nacelle/tools"
)

type ComputerProvider interface {
	Ensure(ctx context.Context, userID string) (*sandbox.Target, error)
	Touch(userID string)
}

type ToolBuilder func(target *sandbox.Target) ([]nacelle.Tool, io.Closer, error)

type BackendFactory func(backend, model string) (nacelle.Backend, error)

type cleanRunner struct{}

func (cleanRunner) env() []string {
	return []string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME")}
}

func (c cleanRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = c.env()
	return cmd.CombinedOutput()
}

func (c cleanRunner) RunInteractive(_ context.Context, cmd *exec.Cmd) error {
	cmd.Env = c.env()
	return cmd.Run()
}

func RemoteVMTools(target *sandbox.Target) ([]nacelle.Tool, io.Closer, error) {
	if target == nil {
		return nil, nil, fmt.Errorf("runner: no computer target")
	}
	return sandbox.RemoteTools(sandbox.ToolsOptions{Target: target, WorkDir: target.Workdir, Runner: cleanRunner{}})
}

func WebFetchTools() ([]nacelle.Tool, error) {
	return tools.WebFetch()
}

type noComputer struct{}

func (noComputer) Ensure(context.Context, string) (*sandbox.Target, error) {
	return nil, fmt.Errorf("no computer provider is configured")
}

func (noComputer) Touch(string) {}

var providerKeyVars =map[string][]string{
	"anthropic":  {"ANTHROPIC_API_KEY"},
	"openai":     {"OPENAI_API_KEY"},
	"google":     {"GEMINI_API_KEY", "GOOGLE_API_KEY"},
	"openrouter": {"OPENROUTER_API_KEY"},
}

func providerKey(backend string) string {
	for _, name := range providerKeyVars[backend] {
		if value := os.Getenv(name); value != "" {
			return value
		}
	}
	return ""
}

func EnvBackends(backend, model string) (nacelle.Backend, error) {
	if _, known := providerKeyVars[backend]; !known {
		return nil, fmt.Errorf("unknown backend %q", backend)
	}
	key := providerKey(backend)
	if key == "" {
		return nil, fmt.Errorf("backend %s has no server key configured", backend)
	}
	cfg := provider.Config{Backend: backend, Model: model}
	if backend != "anthropic" {
		cfg.APIKey = key
	}
	return provider.New(cfg)
}

func EnvSecrets() Redactor {
	var secrets []string
	for _, names := range providerKeyVars {
		for _, name := range names {
			secrets = append(secrets, os.Getenv(name))
		}
	}
	return NewRedactor(secrets...)
}
