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
		suite.AntenneTools(), suite.VisionTools(), suite.EchoTools(), suite.ToileTools(), suite.DocsTools(), suite.FeuillesTools(),
		suite.DiapoTools(), suite.ClapTools())
	for _, section := range mcpSections {
		if !slices.Contains(served, section.tool) {
			t.Errorf("the note keyed on %s never shows: no server exposes it", section.tool)
		}
	}
	fresh := []string{"plume_", "capsule_", "journal_", "antenne_", "vision_", "echo_", "toile_", "docs_", "feuilles_", "diapo_", "clap_"}
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

func TestTheSceneNotesSayHowToWriteADeckAndAVideo(t *testing.T) {
	cases := []struct {
		marker string
		tools  []string
		says   []string
	}{
		{"diapo_list_decks", suite.DiapoTools(), []string{"Always call diapo_format_guide before the first write", "pass the rev you read on every write",
			"Prefer diapo_edit_deck", "diapo_write_deck for new or restructured content", "retry once", "only the assets diapo_get_deck lists",
			"`layout: raw`", "no Date.now, no unseeded random, no network", "only when the user asked", "diapo_delete_deck waits on Diapo's approval card", "Give the user the url"}},
		{"clap_list_projects", suite.ClapTools(), []string{"Always call clap_format_guide before the first write", "pass the rev you read on every write",
			"Prefer clap_edit_project", "clap_write_project for new or restructured content", "retry once", "call clap_list_assets first",
			"`layout: raw`", "no Date.now, no unseeded random, no network", "clap_render only when the user asks", "clap_delete_project waits on Clap's approval card",
			"Give the user the url"}},
	}
	for _, tc := range cases {
		note := mcpNote([]string{tc.marker}, nil, suite.Defaults{})
		for _, want := range tc.says {
			if !strings.Contains(note, want) {
				t.Errorf("the %s note must say %q", tc.marker, want)
			}
		}
		for word := range strings.FieldsFuncSeq(note, func(r rune) bool { return r != '_' && (r < 'a' || r > 'z') }) {
			if strings.Contains(word, "_") && strings.HasPrefix(word, strings.SplitN(tc.marker, "_", 2)[0]+"_") && !slices.Contains(tc.tools, word) {
				t.Errorf("the %s note names %s, which its server does not expose", tc.marker, word)
			}
		}
	}
	if note := mcpNote([]string{"docs_search"}, nil, suite.Defaults{}); strings.Contains(note, "diapo_") || strings.Contains(note, "clap_") {
		t.Fatal("an app that is switched off must leave no note")
	}
}
