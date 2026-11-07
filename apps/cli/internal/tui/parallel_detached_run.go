package tui

import (
	"context"

	"github.com/FacileStudio/nacelle"
)

// defaultConcurrency is the subagent cap a detached fan-out uses when the user
// set none.
const defaultConcurrency = 16

// concurrencyCap returns the user's subagent cap, or the default when unset.
func concurrencyCap(n int) int {
	if n > 0 {
		return n
	}
	return defaultConcurrency
}

// detachedCallbacks routes a detached batch's live tool and spend updates to
// its rows through the subagentUpdates channel.
func detachedCallbacks(id string) nacelle.ParallelSubAgentCallbacks {
	return nacelle.ParallelSubAgentCallbacks{
		Tool: func(b string, i int, t string) { subagentUpdates <- subagentUpdate{batch: id, idx: i, tool: t} },
		LiveUsage: func(b string, i int, u nacelle.Usage) {
			subagentUpdates <- subagentUpdate{batch: id, idx: i, usage: u, spend: true}
		},
		ToolDone: ReportSubagentDone,
	}
}

// runDetached is the fan-out goroutine. It reads only the config snapshot it
// was handed and talks to the update loop through the detached channel alone,
// ending with a done marker so the loop drops the batch's cancel func.
func runDetached(ctx context.Context, cfg nacelle.Config, id string, tasks []string, opts nacelle.ParallelSubAgentOptions) {
	results, err := nacelle.DelegateParallel(ctx, cfg, tasks, opts)
	if err != nil {
		detached <- detachedResult{batch: id, idx: -1, err: err.Error()}
		detached <- detachedResult{batch: id, done: true}
		return
	}
	for next := range results {
		detached <- detachedResult{batch: id, idx: next.Index, result: next.Result, err: next.Err, usage: next.Usage}
	}
	detached <- detachedResult{batch: id, done: true}
}
