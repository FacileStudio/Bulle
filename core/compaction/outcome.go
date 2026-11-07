package compaction

import "github.com/FacileStudio/nacelle"

// Action names what a maintenance step did to the conversation it was handed.
type Action uint8

const (
	// Untouched means the conversation came back as it went in.
	Untouched Action = iota
	// Masked means oversized tool results in the history were stubbed.
	Masked
	// Folded means the history was replaced by the ledger.
	Folded
	// Refused means the fold would have outweighed the turns it replaced.
	Refused
	// Stale means the conversation moved on while the pass ran.
	Stale
)

// Outcome holds the state and result of a compaction pass. Action, Masked and
// Applied say what the pass did once it settled, MaskOnly that the history was
// too small to fold the conversation under its trigger, ThrashHit that this
// pass is the one that reached the thrash limit, and Warning a degraded step
// that did not stop the pass.
type Outcome struct {
	Before      int64
	After       int64
	Plan        []Span
	Fold        Fold
	Tier        Tier
	Summary     string
	Judged      bool
	Consolidate bool
	Stage       string
	Err         error
	Action      Action
	Masked      MicroStats
	Applied     Stats
	MaskOnly    bool
	ThrashHit   bool
	Warning     string
}

// FailedAt names which stage of a compaction pass failed.
func (o Outcome) FailedAt() string {
	if o.Stage != "" {
		return o.Stage
	}
	return "summary"
}

// Installs reports whether a finished pass may replace the conversation.
func (o Outcome) Installs() bool {
	return o.Summary != "" || (o.Judged && o.Fold.LedgerSize() == 0)
}

// Freed reports the token count reduction between before and after sizes.
func (o Outcome) Freed() int64 {
	if o.After <= 0 || o.Before <= o.After {
		return 0
	}
	return o.Before - o.After
}

// Apply reassembles the conversation using the outcome and records after size.
func (o *Outcome) Apply(conv []nacelle.Message) ([]nacelle.Message, Stats) {
	newConv, stats := Apply(conv, o.Plan, o.Summary, o.Fold.Survives, o.Consolidate)
	if !stats.Stale && !stats.Refused {
		o.After = max(o.Before-stats.Before+stats.After, 0)
	} else {
		o.After = o.Before
	}
	return newConv, stats
}
