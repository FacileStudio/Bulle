package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/FacileStudio/nacelle"
)

const (
	maxSearchesPerRun = 20
	maxFetchesPerRun  = 30
	fetchTokenBudget  = 8000
	charsPerToken     = 4
)

// capped bounds one web tool for the length of a run: how many calls it may
// take, and how much text one call may hand back to the model.
type capped struct {
	nacelle.Tool
	left   *atomic.Int64
	budget int
}

// hostTools builds the run's host tools with the web limits applied.
func (r *Runner) hostTools(lr *liveRun) ([]nacelle.Tool, error) {
	host, err := r.cfg.HostTools(lr.userID)
	if err != nil {
		return nil, err
	}
	return capWeb(host), nil
}

// capWeb wraps web_search and web_fetch with their per-run limits and leaves
// every other tool as it is. It is called once per run, so the counts start
// fresh each time.
func capWeb(tools []nacelle.Tool) []nacelle.Tool {
	out := make([]nacelle.Tool, 0, len(tools))
	for _, tool := range tools {
		switch tool.Name() {
		case "web_search":
			out = append(out, newCapped(tool, maxSearchesPerRun, 0))
		case "web_fetch":
			out = append(out, newCapped(tool, maxFetchesPerRun, fetchTokenBudget*charsPerToken))
		default:
			out = append(out, tool)
		}
	}
	return out
}

func newCapped(tool nacelle.Tool, calls int64, budget int) capped {
	left := &atomic.Int64{}
	left.Store(calls)
	return capped{Tool: tool, left: left, budget: budget}
}

func (c capped) Run(ctx context.Context, input json.RawMessage) (string, error) {
	if c.left.Add(-1) < 0 {
		return "", fmt.Errorf("%s has reached its limit for this run; answer with what you already have", c.Name())
	}
	out, err := c.Tool.Run(ctx, input)
	return clip(out, c.budget), err
}

func (c capped) IsReadOnly() bool {
	readOnly, ok := c.Tool.(nacelle.ReadOnlyTool)
	return ok && readOnly.IsReadOnly()
}

func clip(text string, budget int) string {
	if budget <= 0 || len(text) <= budget {
		return text
	}
	kept := strings.ToValidUTF8(text[:budget], "")
	if cut := strings.LastIndexByte(kept, '\n'); cut > budget/2 {
		kept = kept[:cut]
	}
	return kept + fmt.Sprintf("\n\n[truncated at about %d tokens: %d of %d bytes shown]", budget/charsPerToken, len(kept), len(text))
}
