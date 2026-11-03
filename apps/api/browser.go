package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/core/sandbox"
)

// browserBundle is the bulle-browser daemon the API copies into each VM. A
// nil bundle means browsing is off and the VM gets the plain tool set.
type browserBundle struct {
	data   []byte
	logger *slog.Logger
}

func loadBrowserBundle(path string, logger *slog.Logger) (*browserBundle, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading the bulle-browser bundle: %w", err)
	}
	return &browserBundle{data: data, logger: logger}, nil
}

func (b *browserBundle) tools() runner.ToolBuilder {
	if b == nil {
		return runner.RemoteVMTools
	}
	return runner.BrowserVMTools
}

// preflight checks the VM is the one expected, then delivers the bundle.
func (b *browserBundle) preflight(ctx context.Context, target *sandbox.Target) error {
	if target == nil {
		return nil
	}
	if _, err := sandbox.PreflightCheck(ctx, target, sandbox.GuardOptions{
		ExpectedUser:    target.User,
		ExpectedWorkdir: target.Workdir,
		Runner:          runner.CleanRunner{},
	}); err != nil {
		return err
	}
	b.deliver(ctx, target)
	return nil
}

// deliver copies the bundle into the VM when it differs. A failure only
// warns: the browser tools then fail on their own, and the rest of the
// computer keeps working.
func (b *browserBundle) deliver(ctx context.Context, target *sandbox.Target) {
	if b == nil {
		return
	}
	copied, err := sandbox.InstallBrowser(ctx, target, runner.CleanRunner{}, b.data)
	if err != nil {
		b.logger.Warn("failed to install bulle-browser", slog.String("vm", target.Name), slog.Any("error", err))
		return
	}
	if copied {
		b.logger.Info("installed bulle-browser", slog.String("vm", target.Name))
	}
}
