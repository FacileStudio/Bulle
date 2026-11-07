package compaction

import (
	"fmt"
	"strings"
	"testing"

	"github.com/FacileStudio/nacelle"
)

func TestAskWithEmptyPrevious(t *testing.T) {
	got := AskWith("", false)
	if got != CompactAsk {
		t.Errorf("AskWith(\"\", false) = %q, want %q", got, CompactAsk)
	}
	gotConsolidate := AskWith("", true)
	if gotConsolidate != CompactAsk {
		t.Errorf("AskWith(\"\", true) = %q, want %q", gotConsolidate, CompactAsk)
	}
}

func TestAskWithConsolidate(t *testing.T) {
	prev := "Decisions:\n- keep fast path"
	got := AskWith(prev, true)
	wantChunk := fmt.Sprintf(ConsolidateAsk, MaxLedgerTokens)
	if !strings.Contains(got, CompactAsk) || !strings.Contains(got, wantChunk) || !strings.Contains(got, prev) {
		t.Errorf("AskWith(%q, true) missing expected chunks: got %q", prev, got)
	}
}

func TestAskWithKeep(t *testing.T) {
	prev := "Decisions:\n- keep fast path"
	got := AskWith(prev, false)
	if !strings.Contains(got, CompactAsk) || !strings.Contains(got, KeepAsk) || !strings.Contains(got, prev) {
		t.Errorf("AskWith(%q, false) missing expected chunks: got %q", prev, got)
	}
}

func TestPromptRendersTheLedgerBlocksAsOneUserMessage(t *testing.T) {
	conv := []nacelle.Message{
		nacelle.UserText("task"),
		{Role: nacelle.RoleAssistant, Parts: []nacelle.Part{nacelle.ToolCall{ID: "c1", Name: "read_file", Input: []byte(`{"path":"a.go"}`)}}},
		{Role: nacelle.RoleUser, Parts: []nacelle.Part{nacelle.ToolResult{ID: "c1", Name: "read_file", Result: "package a"}}},
		nacelle.UserText("recent"),
	}
	plan := Plan(conv, Policy{KeepTurns: 1, AnchorMessages: 1})
	fold := Fold{Ledger: []Block{{Start: 1, End: 3}}}
	msgs := Prompt(conv, plan, fold, false)
	if len(msgs) != 1 || msgs[0].Role != nacelle.RoleUser || len(msgs[0].Parts) != 1 {
		t.Fatalf("Prompt = %+v, want one user message with one part", msgs)
	}
	text, ok := msgs[0].Parts[0].(nacelle.Text)
	if !ok {
		t.Fatalf("part = %T, want text, never a raw tool part", msgs[0].Parts[0])
	}
	for _, want := range []string{"read_file", "a.go", "package a", CompactAsk} {
		if !strings.Contains(text.Text, want) {
			t.Errorf("prompt is missing %q: %q", want, text.Text)
		}
	}
}

func TestPromptEmptyHistoryDropsLedger(t *testing.T) {
	conv := []nacelle.Message{
		nacelle.UserText("task"),
		BuildLedger("", "Decisions:\n- recorded"),
	}
	plan := Plan(conv, Policy{KeepTurns: 1, AnchorMessages: 1})
	fold := Fold{}
	msgs := Prompt(conv, plan, fold, true)
	if len(msgs) != 1 {
		t.Fatalf("len(msgs) = %d, want 1", len(msgs))
	}
	text := msgs[0].Parts[0].(nacelle.Text).Text
	if strings.Contains(text, "Decisions:") {
		t.Errorf("prompt without history contains ledger text: %q", text)
	}
}

func TestSystemPromptSections(t *testing.T) {
	sections := []string{
		"Decisions", "Constraints", "Plan", "State",
		"Artifacts", "Ruled out", "Open questions",
	}
	for _, sec := range sections {
		if !strings.Contains(SystemPrompt, sec) {
			t.Errorf("SystemPrompt missing required section %q", sec)
		}
	}
}
