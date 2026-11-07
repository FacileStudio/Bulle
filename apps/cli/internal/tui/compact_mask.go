package tui

import "github.com/FacileStudio/Bulle/core/compaction"

// takeStep takes in a step the engine finished on the UI thread, a soft pass or
// a smart pass that never needed the summarizer, and says what it did.
func (m *Model) takeStep(outcome compaction.Outcome) {
	m.report(outcome)
	m.sayThrash(outcome)
}

// report books what a settled step did to the session and says it. A step
// that left the conversation as it was says nothing, except a fold that went
// stale, which says why nothing was folded.
func (m *Model) report(outcome compaction.Outcome) {
	switch outcome.Action {
	case compaction.Masked:
		m.countMask(outcome)
		m.say(fromCompact, maskReport(outcome, m.last))
	case compaction.Folded:
		m.installFold(outcome)
	case compaction.Refused:
		m.last = compacted{tier: outcome.Tier}
		m.say(fromCompact, "context unchanged — the ledger would have outweighed the turns it folds")
	case compaction.Stale:
		if outcome.Err == nil && outcome.Installs() {
			m.say(fromCompact, "the conversation changed while the pass ran — nothing was folded")
		}
	}
	if outcome.Warning != "" {
		m.say(fromCompact, "compaction "+outcome.Warning)
	}
}

// countMask debits the session size by the stub pass's estimate and books the
// stubbed results. The debit is deliberately an estimate: the next sized()
// overwrites it with the backend's own count.
func (m *Model) countMask(outcome compaction.Outcome) {
	m.trimmed += outcome.Masked.Results
	m.size = outcome.After
	start, end, _ := compaction.HistoryRange(outcome.Plan)
	m.last = compacted{evictCut: end - start, kept: activeLen(outcome.Plan), results: outcome.Masked.Results, tier: outcome.Tier}
}

// sayFailure names a summarizing pass that failed or came back empty, and
// whether the stub fallback stood in for it.
func (m *Model) sayFailure(outcome compaction.Outcome) {
	masked := outcome.Action == compaction.Masked
	switch {
	case outcome.Err != nil:
		m.say(fromCompact, "compaction "+outcome.FailedAt()+" failed · "+outcome.Err.Error()+maskNote(masked))
	case !outcome.Installs():
		m.say(fromCompact, "compaction summary came back empty"+maskNote(masked))
	}
}

func (m *Model) sayThrash(outcome compaction.Outcome) {
	if outcome.ThrashHit && m.compactAt > 0 {
		m.say(fromClient, "compaction keeps leaving the context over the threshold — one very large result is likeliest; /clear, read in chunks, then /compact")
	}
}

// maskNote is what the mask managed, appended to a notice that had to mention
// it: a fallback that did not land must not claim it did, so the one line that
// says "masked instead" is the one place that has to ask.
func maskNote(masked bool) string {
	if masked {
		return " — masked instead"
	}
	return " — nothing was masked: the conversation changed while the pass ran"
}

func activeLen(plan []compaction.Span) int {
	total := 0
	for _, span := range plan {
		if span.Zone == compaction.ZoneActive {
			total += span.End - span.Start
		}
	}
	return total
}
