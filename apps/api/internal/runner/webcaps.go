package runner

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/FacileStudio/nacelle"
)

// webCalls is how many times one run may call each of sonar's tools. sonar
// counts searches against the user's daily quota; this keeps one runaway run
// from spending all of it.
var webCalls = map[string]int64{"sonar_search": 20, "sonar_fetch": 30}

// capWeb wraps sonar_search and sonar_fetch with their per-run limits and
// leaves every other tool as it is. It is called once per run, so the counts
// start fresh each time.
func capWeb(tools []nacelle.Tool) []nacelle.Tool {
	out := make([]nacelle.Tool, 0, len(tools))
	for _, tool := range tools {
		if calls, ok := webCalls[tool.Spec().Name]; ok {
			tool = newCapped(tool, calls)
		}
		out = append(out, tool)
	}
	return out
}

// newCapped bounds how many calls one tool may take for the length of a run.
func newCapped(tool nacelle.Tool, calls int64) nacelle.Tool {
	left := &atomic.Int64{}
	left.Store(calls)
	name := tool.Spec().Name
	return nacelle.Wrap(tool, nil, func(next nacelle.RunFunc) nacelle.RunFunc {
		return func(ctx context.Context, call nacelle.Call, emit func(string)) (nacelle.Result, error) {
			if left.Add(-1) < 0 {
				return nacelle.Result{}, fmt.Errorf("%s has reached its limit for this run; answer with what you already have", name)
			}
			return next(ctx, call, emit)
		}
	})
}
