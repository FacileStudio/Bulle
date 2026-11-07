package runner

import (
	"context"

	"github.com/FacileStudio/Bulle/core/compaction"
	"github.com/FacileStudio/Bulle/core/engine"
	"github.com/FacileStudio/Bulle/core/settings"
	"github.com/FacileStudio/nacelle"
)

// maintain keeps the conversation inside the model's window, measured with the
// agent's own backend so the system prompt and the tool schemas count, and
// saves the result when a pass changed it. overflow forces a pass after the
// provider refused the conversation as too long. It answers the conversation
// to send.
func (r *Runner) maintain(ctx context.Context, lr *liveRun, agent *nacelle.Agent, messages []nacelle.Message, overflow bool) []nacelle.Message {
	backend := agent.Backend()
	compactor := compaction.NewEngine(serverPolicy(backend.Capabilities().ContextWindow), nil, compaction.BackendAgent(backend))
	size := measure(ctx, agent, messages)
	next, report := compactor.Maintain(ctx, messages, size, overflow)
	if report.Action != compaction.Masked && report.Action != compaction.Folded {
		return messages
	}
	next = engine.Repair(next)
	r.cfg.Logger.Info("compacted the conversation", "run", lr.id, "before", report.Before, "after", report.After, "folded", report.Action == compaction.Folded, "warning", report.Warning)
	if err := r.saveHistory(context.WithoutCancel(ctx), lr.session.ID, next); err != nil {
		r.cfg.Logger.Error("saving the compacted history", "run", lr.id, "error", err)
	}
	return next
}

// serverPolicy is the shipped ladder against the model's window, with the
// windowless fallback ceiling when the backend reports none.
func serverPolicy(window int64) compaction.Policy {
	policy := compaction.Policy{Window: window}
	if window <= 0 {
		policy.Ceiling = settings.DefaultCompactAt
	}
	return policy.Resolved()
}

// measure is what the next request would cost: the backend's own count when
// it offers one, otherwise the byte estimate of the conversation.
func measure(ctx context.Context, agent *nacelle.Agent, messages []nacelle.Message) int64 {
	if count, err := agent.CountTokens(ctx, messages); err == nil && count > 0 {
		return count
	}
	return compaction.EstTokens(compaction.Bytes(messages))
}
