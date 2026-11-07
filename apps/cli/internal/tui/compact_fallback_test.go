package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/core/compaction"
	"github.com/FacileStudio/nacelle"
)

// The mask refuses a plan that no longer covers the conversation, the same as
// the fold does, and the notice it feeds says nothing landed rather than
// promising a masking that did not happen: tombstoning whatever now sits at
// those indices would be an edit to a conversation the pass never measured.
func TestMaskFallbackRefusesAStalePlan(t *testing.T) {
	m := sized()
	m.conversation = bigConversation()
	m.size = compactAt + 25_000
	before := m.size
	gone := []nacelle.Message{nacelle.UserText("a conversation that is gone"), nacelle.AssistantText("and its answer")}

	m.settleCompaction(compaction.Outcome{Before: before, Plan: compaction.Plan(gone, m.engine().Policy), Tier: compaction.Smart, Err: errors.New("summarizer hiccuped")})

	if m.trimmed != 0 || m.size != before {
		t.Errorf("a stale mask changed the session: trimmed = %d, size = %d", m.trimmed, m.size)
	}
	if said := strings.Join(spoken(m), " "); !strings.Contains(said, "nothing was masked") {
		t.Errorf("said = %q, want it to report that no fallback landed", said)
	}
}

func TestSettleCompactionFallsBackToTheMaskOnFailure(t *testing.T) {
	m := sized()
	m.conversation = bigConversation()
	m.size = compactAt + 25_000
	outcome := compaction.Outcome{Before: m.size, Plan: m.plan(), Tier: compaction.Smart, Err: errors.New("summarizer hiccuped")}

	m.settleCompaction(outcome)

	if m.compacting {
		t.Error("compacting still true after the fallback")
	}
	if len(m.conversation) != len(bigConversation()) {
		t.Errorf("mask fallback changed the message count: %d", len(m.conversation))
	}
	saved := 0
	for _, message := range m.conversation {
		for _, part := range message.Parts {
			result, ok := part.(nacelle.ToolResult)
			if ok && strings.HasPrefix(result.Result, droppedNotice) {
				saved++
			}
		}
	}
	if saved == 0 {
		t.Error("a failed summary bought no headroom: nothing was masked")
	}
	said := strings.Join(spoken(m), " ")
	if !strings.Contains(said, "compaction summary failed") || !strings.Contains(said, "masked instead") {
		t.Errorf("failure not reported: %v", said)
	}
	if !strings.Contains(said, "✂ Compaction summary") {
		t.Errorf("mask fallback did not report: %v", said)
	}
}
