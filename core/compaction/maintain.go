package compaction

import (
	"context"

	"github.com/FacileStudio/nacelle"
)

// Maintain is the one entry point that keeps a conversation inside its budget.
// size is what the conversation costs as the backend counts it, and overflow
// says the backend has just refused it as too long.
//
// It picks the tier from size, runs the model-free stub pass at the soft tier
// behind the cache guard (MinCleared), and at the smart tier runs the judged
// fold, stubbing instead when the fold cannot land or fails. The thrash guard
// stands it down after repeated passes that did not land, unless overflow
// forces one more. conv is never modified: the result is a new conversation.
func (e *Engine) Maintain(ctx context.Context, conv []nacelle.Message, size int64, overflow bool) ([]nacelle.Message, Outcome) {
	conv = Snapshot(conv)
	switch e.due(size, overflow) {
	case Soft:
		return conv, e.SoftPass(conv, size)
	case Smart:
		outcome, pending := e.Prepare(conv, size)
		if !pending {
			return conv, outcome
		}
		return e.Settle(conv, e.RunPassWithSize(ctx, conv, size, overflow))
	default:
		return conv, Outcome{Before: size, After: size, Tier: Below}
	}
}

func (e *Engine) due(size int64, overflow bool) Tier {
	if overflow {
		return Smart
	}
	if !e.Policy.Over(size) || e.Thrashed() {
		return Below
	}
	return e.Policy.Tier(size)
}

// SoftPass stubs the oversized tool results in conv's history, in place, when
// doing so frees at least MinCleared bytes. Under that floor the cached prefix
// the stubs would invalidate is worth more than the context they free, so the
// conversation is left untouched.
func (e *Engine) SoftPass(conv []nacelle.Message, size int64) Outcome {
	outcome := Outcome{Before: size, After: size, Plan: Plan(conv, e.Policy), Tier: Soft}
	if DroppableBytes(conv, outcome.Plan) < MinCleared {
		return outcome
	}
	outcome = mask(conv, outcome)
	if outcome.Masked.Results == 0 {
		outcome.Action = Untouched
	}
	return outcome
}

// Prepare is the model-free half of a smart pass. It reports pending when the
// summarizing pass should run. Otherwise the outcome is final: a conversation
// with no history is untouched, and one whose history is too small to fold it
// under the trigger is stubbed in place, which counts toward the thrash guard.
func (e *Engine) Prepare(conv []nacelle.Message, size int64) (Outcome, bool) {
	outcome := Outcome{Before: size, After: size, Plan: Plan(conv, e.Policy), Tier: e.Policy.Tier(size)}
	if _, _, ok := HistoryRange(outcome.Plan); !ok {
		return outcome, false
	}
	if CanEvictionLandUnder(size, conv, outcome.Plan, e.Policy) {
		return outcome, true
	}
	outcome.MaskOnly = true
	return e.thrash(mask(conv, outcome)), false
}

func mask(conv []nacelle.Message, outcome Outcome) Outcome {
	outcome.Masked = Tombstone(conv, outcome.Plan)
	outcome.After = max(outcome.Before-EstTokens(outcome.Masked.Bytes), 0)
	outcome.Action = Masked
	return outcome
}

func (e *Engine) thrash(outcome Outcome) Outcome {
	outcome.ThrashHit = e.CheckThrash(outcome.After) && e.ThrashCount() == e.limit()
	return outcome
}
