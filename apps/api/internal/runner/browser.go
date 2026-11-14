package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
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

// browserTyping is the approval policy of browser_type in a run third-party
// text reached: the owner sees the text the page would get and the origin of
// the page it goes to, since a page's script can send what is typed without
// a request the daemon holds. The page's origin is the scope, so a grant
// never covers typing on another origin.
type browserTyping struct {
	vm *lazyVM
}

type typeInput struct {
	Text   string `json:"text"`
	Submit bool   `json:"submit"`
}

func (b browserTyping) describe(ctx context.Context, input json.RawMessage) (map[string]any, string, error) {
	var in typeInput
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, "", err
	}
	origin := b.origin(ctx)
	record := map[string]any{"text": in.Text, "origin": origin}
	if in.Submit {
		record["submit"] = true
	}
	if origin == "unknown" {
		origin = ""
	}
	return record, origin, nil
}

func (b browserTyping) denied(context.Context, json.RawMessage) {}

// origin is the scheme and host of the page the browser has open, or
// "unknown" when the browser cannot say.
func (b browserTyping) origin(ctx context.Context) string {
	browser, err := b.vm.browser(ctx)
	if err != nil {
		return "unknown"
	}
	status, err := browser.Browser(ctx, "status", struct{}{})
	if err != nil {
		return "unknown"
	}
	page, err := url.Parse(status.URL)
	if err != nil || page.Host == "" {
		return "unknown"
	}
	return page.Scheme + "://" + page.Host
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
