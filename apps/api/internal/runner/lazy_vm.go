package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/FacileStudio/Bulle/core/sandbox"
	"github.com/FacileStudio/nacelle"
)

type lazyVM struct {
	runner *Runner
	user   string

	mu      sync.Mutex
	done    bool
	browsed bool
	tools   map[string]nacelle.Tool
	closer  io.Closer
	err     error
}

func (r *Runner) lazyVMTools(user string, g *gate) ([]nacelle.Tool, *lazyVM, error) {
	probe, closer, err := r.cfg.VMTools(&sandbox.Target{})
	if err != nil {
		return nil, nil, fmt.Errorf("opening your computer: %w", err)
	}
	if closer != nil {
		_ = closer.Close()
	}
	vm := &lazyVM{runner: r, user: user}
	g.addPolicy(browserConfirmTool, browserConfirm{vm: vm})
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
	v.browsed = v.browsed || strings.HasPrefix(name, "browser_")
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
		v.abortBrowser()
		_ = v.closer.Close()
		v.closer = nil
	}
}

// browser is the daemon client of the run's VM session, booting the VM if
// nothing has yet.
func (v *lazyVM) browser(ctx context.Context) (sandbox.Browser, error) {
	if _, err := v.open(ctx, browserConfirmTool); err != nil {
		return nil, err
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	browser, ok := v.closer.(sandbox.Browser)
	if !ok {
		return nil, fmt.Errorf("this computer has no browser")
	}
	return browser, nil
}

// abortBrowser drops every held request and stops page loads when a run that
// used the browser ends, however it ended: finished, stopped or timed out.
func (v *lazyVM) abortBrowser() {
	browser, ok := v.closer.(sandbox.Browser)
	if !v.browsed || !ok {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), browserAbortWait)
	defer cancel()
	if _, err := browser.Browser(ctx, "abort", struct{}{}); err != nil {
		v.runner.cfg.Logger.Warn("aborting the browser", "user", v.user, "error", err)
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

// RunImages forwards to the VM tool, keeping the pictures an image tool such
// as browser_screenshot returns.
func (t lazyTool) RunImages(ctx context.Context, input json.RawMessage) (string, []nacelle.Image, error) {
	tool, err := t.vm.open(ctx, t.Name())
	if err != nil {
		return "", nil, err
	}
	if images, ok := tool.(nacelle.ImageTool); ok {
		return images.RunImages(ctx, input)
	}
	result, err := tool.Run(ctx, input)
	return result, nil, err
}
