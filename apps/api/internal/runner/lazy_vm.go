package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/FacileStudio/Bulle/core/sandbox"
	"github.com/FacileStudio/nacelle"
)

type lazyVM struct {
	runner *Runner
	user   string

	mu     sync.Mutex
	done   bool
	tools  map[string]nacelle.Tool
	closer io.Closer
	err    error
}

func (r *Runner) lazyVMTools(user string) ([]nacelle.Tool, *lazyVM, error) {
	probe, closer, err := r.cfg.VMTools(&sandbox.Target{})
	if err != nil {
		return nil, nil, fmt.Errorf("opening your computer: %w", err)
	}
	if closer != nil {
		_ = closer.Close()
	}
	vm := &lazyVM{runner: r, user: user}
	wrapped := make([]nacelle.Tool, len(probe))
	for i, tool := range probe {
		wrapped[i] = lazyTool{Tool: tool, vm: vm}
	}
	return wrapped, vm, nil
}

func (v *lazyVM) open(ctx context.Context, name string) (nacelle.Tool, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if !v.done {
		v.done = true
		v.tools, v.closer, v.err = v.boot(ctx)
	}
	if v.err != nil {
		return nil, v.err
	}
	tool, ok := v.tools[name]
	if !ok {
		return nil, fmt.Errorf("opening your computer: tool %q is missing", name)
	}
	return tool, nil
}

func (v *lazyVM) boot(ctx context.Context) (map[string]nacelle.Tool, io.Closer, error) {
	cfg := v.runner.cfg
	target, err := cfg.Computers.Ensure(ctx, v.user)
	if err != nil {
		return nil, nil, fmt.Errorf("your computer is not available: %w", err)
	}
	if cfg.Preflight != nil {
		if err := cfg.Preflight(ctx, target); err != nil {
			return nil, nil, fmt.Errorf("opening your computer: %w", err)
		}
	}
	tools, closer, err := cfg.VMTools(target)
	if err != nil {
		return nil, nil, fmt.Errorf("opening your computer: %w", err)
	}
	return nacelle.ToolsByName(tools), closer, nil
}

func (v *lazyVM) close() {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.closer != nil {
		_ = v.closer.Close()
		v.closer = nil
	}
}

type lazyTool struct {
	nacelle.Tool
	vm *lazyVM
}

func (t lazyTool) Run(ctx context.Context, input json.RawMessage) (string, error) {
	tool, err := t.vm.open(ctx, t.Name())
	if err != nil {
		return "", err
	}
	return tool.Run(ctx, input)
}
