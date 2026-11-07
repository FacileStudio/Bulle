package compaction

import "github.com/FacileStudio/nacelle"

// Settle turns a finished pass into the conversation that replaces conv. A
// pass that installs folds the history into the ledger, unless the fold went
// stale or would have outweighed what it replaced. A pass that failed or came
// back empty stubs the history in place instead, so the attempt still frees the
// bulky tool output, unless conv no longer matches the plan the pass measured.
// Every settled pass counts toward the thrash guard.
func (e *Engine) Settle(conv []nacelle.Message, outcome Outcome) ([]nacelle.Message, Outcome) {
	if outcome.Err != nil || !outcome.Installs() {
		return conv, e.thrash(fallback(conv, outcome))
	}
	next, outcome := install(conv, outcome)
	return next, e.thrash(outcome)
}

func fallback(conv []nacelle.Message, outcome Outcome) Outcome {
	if !Covers(conv, outcome.Plan) {
		outcome.After = outcome.Before
		outcome.Action = Stale
		return outcome
	}
	return mask(conv, outcome)
}

func install(conv []nacelle.Message, outcome Outcome) ([]nacelle.Message, Outcome) {
	next, stats := outcome.Apply(conv)
	outcome.Applied = stats
	switch {
	case stats.Stale:
		outcome.Action = Stale
		return conv, outcome
	case stats.Refused:
		outcome.Action = Refused
		return conv, outcome
	}
	outcome.Action = Folded
	return next, outcome
}
