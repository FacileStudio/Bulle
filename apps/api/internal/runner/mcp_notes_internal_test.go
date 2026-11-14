package runner

import (
	"slices"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
)

func TestEveryMCPNoteIsKeyedOnAServedTool(t *testing.T) {
	served := slices.Concat(suite.CourrierTools(), suite.AgendaTools(), suite.SonarTools(), suite.NuageTools(),
		suite.OpusTools(), suite.SablierTools(), suite.GloutonTools(), suite.PlumeTools(), suite.CapsuleTools(), suite.JournalTools(),
		suite.AntenneTools(), suite.VisionTools(), suite.EchoTools(), suite.ToileTools(), suite.DocsTools(), suite.FeuillesTools())
	for _, section := range mcpSections {
		if !slices.Contains(served, section.tool) {
			t.Errorf("the note keyed on %s never shows: no server exposes it", section.tool)
		}
	}
	fresh := []string{"plume_", "capsule_", "journal_", "antenne_", "vision_", "echo_", "toile_", "docs_", "feuilles_"}
	for _, name := range untrustedSources {
		app := slices.IndexFunc(fresh, func(prefix string) bool { return strings.HasPrefix(name, prefix) })
		if app >= 0 && !slices.Contains(served, name) {
			t.Errorf("%s taints a run but no server exposes it", name)
		}
	}
}

func TestANoteSaysAServerConfirmsOnlyTheToolsItDoes(t *testing.T) {
	var defaults suite.Defaults
	for _, section := range mcpSections {
		for sentence := range strings.SplitSeq(section.text(defaults), ". ") {
			if strings.Contains(sentence, "approval card") && !strings.Contains(sentence, "without") {
				checkConfirmedSentence(t, section.tool, sentence)
			}
		}
	}
}

func checkConfirmedSentence(t *testing.T, note, sentence string) {
	t.Helper()
	named := 0
	for word := range strings.FieldsFuncSeq(sentence, func(r rune) bool { return r != '_' && (r < 'a' || r > 'z') }) {
		if !listedTools[word] || readName(word) {
			continue
		}
		named++
		if !serverConfirms(word) {
			t.Errorf("the %s note says %s waits on its server's card: %q", note, word, sentence)
		}
	}
	if named == 0 {
		t.Errorf("the %s note names no tool its server confirms: %q", note, sentence)
	}
}

func TestNoNoteSaysAWriteSkipsEveryCard(t *testing.T) {
	var defaults suite.Defaults
	for _, section := range mcpSections {
		if text := section.text(defaults); strings.Contains(text, "at once without") || strings.Contains(text, "without a card") {
			t.Errorf("the %s note says a write never waits on a card, though Bulle asks in a tainted chat", section.tool)
		}
	}
	if note := mcpNote([]string{"opus_create_task"}, nil, defaults); !strings.HasPrefix(note, taintedWritesNote) {
		t.Fatal("a mounted app's note must open with how Bulle asks before writes in a tainted chat")
	}
	if mcpNote(nil, nil, defaults) != "" {
		t.Fatal("no mounted app must leave no note")
	}
}
