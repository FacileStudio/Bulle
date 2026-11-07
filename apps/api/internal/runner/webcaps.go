package runner

import (
	"context"
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
		switch tool.Spec().Name {
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

// newCapped bounds one web tool for the length of a run: how many calls it
// may take, and how much text one call may hand back to the model.
func newCapped(tool nacelle.Tool, calls int64, budget int) nacelle.Tool {
	left := &atomic.Int64{}
	left.Store(calls)
	name := tool.Spec().Name
	return nacelle.Wrap(tool, nil, func(next nacelle.RunFunc) nacelle.RunFunc {
		return func(ctx context.Context, call nacelle.Call, emit func(string)) (nacelle.Result, error) {
			if left.Add(-1) < 0 {
				return nacelle.Result{}, fmt.Errorf("%s has reached its limit for this run; answer with what you already have", name)
			}
			out, err := next(ctx, call, emit)
			out.Text = clip(out.Text, budget)
			return out, err
		}
	})
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
