package compaction

import (
	"slices"

	"github.com/FacileStudio/nacelle"
)

// Snapshot creates an isolated clone of a conversation and its message parts.
func Snapshot(conv []nacelle.Message) []nacelle.Message {
	out := slices.Clone(conv)
	for i := range out {
		out[i].Parts = slices.Clone(out[i].Parts)
	}
	return out
}

// CanEvictionLandUnder reports whether evicting the whole history, leaving a
// full ledger in its place, can bring size back to the policy's trigger.
func CanEvictionLandUnder(size int64, conv []nacelle.Message, plan []Span, policy Policy) bool {
	start, end, ok := HistoryRange(plan)
	if !ok {
		return false
	}
	freed := EstTokens(Bytes(conv[start:end]))
	return !policy.Over(size - freed + MaxLedgerTokens)
}
