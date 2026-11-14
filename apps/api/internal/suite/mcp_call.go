package suite

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/FacileStudio/nacelle"
	"github.com/FacileStudio/nacelle/mcp/client"
)

// mcpReadLimit bounds one call to a read-only MCP tool. A read never waits
// for a confirmation, so a server silent this long is stuck, not asking.
const mcpReadLimit = 60 * time.Second

const (
	maxToolImages     = 4
	maxToolImageBytes = 1 << 20
)

// SetLogger sends the line every MCP call writes to logger, the app's own,
// so it reaches Journal with the rest of Bulle's logs. Call it before the
// first run.
func (m *MCP) SetLogger(logger *slog.Logger) {
	if m != nil && logger != nil {
		m.logger = logger
	}
}

// bound names a server's tool as the server does, since suite servers
// already prefix their tools with the app, and runs it as userID whatever
// context the loop calls it with. A read gets mcpReadLimit, a call the server
// may confirm keeps mcpCallLimit, and every call writes one log line.
func (m *MCP) bound(tool nacelle.Tool, app string, userID int64) nacelle.Tool {
	limit := mcpCallLimit
	if tool.Spec().ReadOnly {
		limit = mcpReadLimit
	}
	return nacelle.Wrap(tool, func(spec nacelle.Spec) nacelle.Spec {
		spec.Name = mcpName(spec.Name, app)
		return spec
	}, func(next nacelle.RunFunc) nacelle.RunFunc {
		return func(ctx context.Context, call nacelle.Call, emit func(string)) (nacelle.Result, error) {
			ctx, cancel := context.WithTimeout(WithUser(ctx, userID), limit)
			defer cancel()
			started := m.now()
			result, err := throttled(ctx, func() (nacelle.Result, error) { return next(ctx, call, emit) })
			m.audit(mcpCall{app: app, tool: mcpName(tool.Spec().Name, app), user: userID, args: call.Input}, m.now().Sub(started), err)
			return result, err
		}
	})
}

type mcpCall struct {
	app, tool string
	user      int64
	args      []byte
}

// audit writes the one line a call leaves behind: who called which tool,
// how long it took and how it ended, and a hash of the arguments, never the
// arguments themselves.
func (m *MCP) audit(call mcpCall, took time.Duration, err error) {
	sum := sha256.Sum256(call.args)
	attrs := []any{
		"app", call.app, "tool", call.tool, "user", call.user, "duration_ms", took.Milliseconds(),
		"outcome", outcome(err), "args_sha256", hex.EncodeToString(sum[:8]),
	}
	if err != nil {
		m.logger.Warn("mcp call", attrs...)
		return
	}
	m.logger.Info("mcp call", attrs...)
}

func outcome(err error) string {
	switch {
	case err == nil:
		return "ok"
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	case errors.Is(err, context.Canceled):
		return "canceled"
	}
	return "error"
}

// clipped caps what a tool hands the model, as Bulle's own tools are capped,
// so one large answer cannot fill the context: the text at maxToolOutput,
// the images at the first maxToolImages that fit in maxToolImageBytes
// together, with a note naming how many were left out.
func clipped(tool nacelle.Tool) nacelle.Tool {
	return nacelle.Wrap(tool, nil, func(next nacelle.RunFunc) nacelle.RunFunc {
		return func(ctx context.Context, call nacelle.Call, emit func(string)) (nacelle.Result, error) {
			result, err := next(ctx, call, emit)
			result.Text = clip(result.Text)
			images, dropped := keptImages(result.Images)
			result.Images = images
			if dropped > 0 {
				result.Text += fmt.Sprintf("\n[%d images left out: at most %d images and %d KiB per call]",
					dropped, maxToolImages, maxToolImageBytes>>10)
			}
			return result, err
		}
	})
}

func keptImages(images []nacelle.Image) ([]nacelle.Image, int) {
	var kept []nacelle.Image
	size := 0
	for _, image := range images {
		if len(kept) == maxToolImages || size+len(image.Data) > maxToolImageBytes {
			break
		}
		kept = append(kept, image)
		size += len(image.Data)
	}
	return kept, len(images) - len(kept)
}

// throttled sends a call a server refused with 429 once more, after the wait
// the server asked for, when nacelle says the server refused it before it ran.
func throttled(ctx context.Context, run func() (nacelle.Result, error)) (nacelle.Result, error) {
	result, err := run()
	refusal, ok := errors.AsType[*client.RateLimitError](err)
	if !ok || !refusal.Retryable() {
		return result, err
	}
	select {
	case <-ctx.Done():
		return result, err
	case <-time.After(max(refusal.RetryAfter, time.Second)):
	}
	return run()
}
