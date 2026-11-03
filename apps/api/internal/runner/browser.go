package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/FacileStudio/Bulle/core/sandbox"
	"github.com/FacileStudio/nacelle"
)

const (
	browserConfirmTool = "browser_confirm"
	browserAbortWait   = 15 * time.Second
)

// BrowserVMTools is RemoteVMTools plus the browser_* tools that drive the
// bulle-browser daemon in the VM.
func BrowserVMTools(target *sandbox.Target) ([]nacelle.Tool, io.Closer, error) {
	if target == nil {
		return nil, nil, fmt.Errorf("runner: no computer target")
	}
	return sandbox.RemoteTools(sandbox.ToolsOptions{Target: target, WorkDir: target.Workdir, Runner: CleanRunner{}, Browser: true})
}

// browserConfirm is the approval policy of browser_confirm. The owner sees
// the daemon's own record of the held request, looked up by held_id, never
// the model's arguments. One approval covers the request's origin for the
// rest of the run, and a denial drops the held request at once. Secret-looking
// body fields are redacted on the card.
type browserConfirm struct {
	vm *lazyVM
}

type heldInput struct {
	HeldID string `json:"held_id"`
}

func (b browserConfirm) describe(ctx context.Context, input json.RawMessage) (map[string]any, string, error) {
	var in heldInput
	if err := json.Unmarshal(input, &in); err != nil || in.HeldID == "" {
		return nil, "", fmt.Errorf("browser_confirm needs a held_id")
	}
	browser, err := b.vm.browser(ctx)
	if err != nil {
		return nil, "", err
	}
	status, err := browser.Browser(ctx, "status", struct{}{})
	if err != nil {
		return nil, "", err
	}
	held, ok := status.Find(in.HeldID)
	if !ok {
		return nil, "", fmt.Errorf("no held request %s", in.HeldID)
	}
	return heldRecord(held, b.vm.runner.cfg.Redactor), held.Origin, nil
}

func (b browserConfirm) denied(ctx context.Context, input json.RawMessage) {
	var in heldInput
	if json.Unmarshal(input, &in) != nil || in.HeldID == "" {
		return
	}
	browser, err := b.vm.browser(ctx)
	if err == nil {
		_, err = browser.Browser(ctx, "cancel", in)
	}
	if err != nil {
		b.vm.runner.cfg.Logger.Warn("dropping a denied browser request", "held_id", in.HeldID, "error", err)
	}
}

func heldRecord(held sandbox.HeldRequest, redact Redactor) map[string]any {
	return map[string]any{
		"method":       held.Method,
		"url":          redact.Apply(held.URL),
		"body_preview": redactBody(held.BodyPreview, redact),
		"body_bytes":   held.BodyBytes,
		"held_id":      held.ID,
		"scope":        "Approving also lets this run send other requests to " + held.Origin,
	}
}
