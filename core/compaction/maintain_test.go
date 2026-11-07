package compaction

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/FacileStudio/nacelle"
)

func heavyHistory(results int) []nacelle.Message {
	conv := []nacelle.Message{nacelle.UserText("the task")}
	for i := range results {
		id := fmt.Sprintf("c%d", i)
		conv = append(conv, callMessage(id, "read"), resultMessage(id, "read", strings.Repeat("x", 4*MinResult)))
	}
	return append(conv, nacelle.AssistantText("done"), nacelle.UserText("next"))
}

func TestMaintainLeavesAConversationUnderItsTriggerAlone(t *testing.T) {
	conv := heavyHistory(4)
	engine := NewEngine(Policy{Ceiling: 1_000_000}.Resolved(), nil, BackendAgent(summarizingBackend{answer: "Decisions:\n- x"}))
	out, report := engine.Maintain(context.Background(), conv, 10_000, false)
	if report.Action != Untouched || len(out) != len(conv) {
		t.Fatalf("Maintain under the trigger = %v, %d messages, want untouched", report.Action, len(out))
	}
}

func TestMaintainFoldsAHistoryOverTheSmartTier(t *testing.T) {
	conv := heavyHistory(20)
	size := EstTokens(Bytes(conv))
	policy := Policy{Window: size, Ratios: Ratios{Soft: 0.15, Smart: 0.2}, KeepTokens: 1}.Resolved()
	engine := NewEngine(policy, nil, BackendAgent(summarizingBackend{answer: "Decisions:\n- read everything"}))
	out, report := engine.Maintain(context.Background(), conv, size, false)
	if report.Action != Folded {
		t.Fatalf("Action = %v (err %v), want a fold", report.Action, report.Err)
	}
	if !strings.Contains(LedgerText(out, Plan(out, policy)), "read everything") {
		t.Errorf("the ledger does not carry the summary: %v", out)
	}
	if len(conv) != len(heavyHistory(20)) || !strings.HasPrefix(conv[2].Parts[0].(nacelle.ToolResult).Result, "x") {
		t.Error("Maintain modified the conversation it was handed")
	}
}

func TestMaintainStubsWhenTheSummarizerFails(t *testing.T) {
	conv := heavyHistory(20)
	size := EstTokens(Bytes(conv))
	policy := Policy{Window: size, Ratios: Ratios{Soft: 0.15, Smart: 0.2}, KeepTokens: 1}.Resolved()
	engine := NewEngine(policy, nil, func() (*nacelle.Agent, error) { return nil, errors.New("no model") })
	_, report := engine.Maintain(context.Background(), conv, size, false)
	if report.Err == nil || report.Action != Masked || report.Masked.Results == 0 {
		t.Fatalf("report = %+v, want the failure reported and the history stubbed", report)
	}
	if report.After >= report.Before {
		t.Errorf("After = %d, want under Before %d", report.After, report.Before)
	}
}

func TestMaintainStandsDownWhenThrashedUnlessForced(t *testing.T) {
	conv := heavyHistory(20)
	size := EstTokens(Bytes(conv))
	engine := NewEngine(Policy{Ceiling: 100}.Resolved(), nil, nil)
	engine.SetThrash(DefaultThrashLimit)
	if _, report := engine.Maintain(context.Background(), conv, size, false); report.Action != Untouched {
		t.Fatalf("a thrashed engine ran a pass: %v", report.Action)
	}
	if _, report := engine.Maintain(context.Background(), conv, size, true); report.Action == Untouched {
		t.Fatal("an overflow did not force a pass past the thrash guard")
	}
}

func TestSoftPassHoldsTheCacheGuard(t *testing.T) {
	conv := heavyHistory(1)
	engine := NewEngine(Policy{Ceiling: 100}.Resolved(), nil, nil)
	if report := engine.SoftPass(conv, 5_000); report.Action != Untouched {
		t.Fatalf("a pass under MinCleared stubbed: %+v", report)
	}
}

func TestLandsUnderWithCountsTheOverhead(t *testing.T) {
	conv := judgeSample()
	plan := judgePlan(conv)
	fold := Fold{Ledger: Blocks(conv, plan)}
	trigger := EstTokens(estimateRebuilt(conv, plan, fold)) + 100
	if !LandsUnderWith(conv, plan, fold, trigger, 0) {
		t.Fatal("the fold does not land without overhead")
	}
	if LandsUnderWith(conv, plan, fold, trigger, 200) {
		t.Error("the fold lands although the system prompt and tools push it over the trigger")
	}
}

func TestStubFreesImagesAndKeepsNothingHeavy(t *testing.T) {
	shot := nacelle.ToolResult{ID: "c1", Name: "screenshot", Result: "ok", Images: []nacelle.Image{{MediaType: "image/png", Data: []byte{1}}}}
	conv := []nacelle.Message{
		nacelle.UserText("the task"),
		callMessage("c1", "screenshot"),
		{Role: nacelle.RoleUser, Parts: []nacelle.Part{shot}},
		nacelle.AssistantText("done"),
		nacelle.UserText("next"),
	}
	if PartBytes(shot) < mediaBytes {
		t.Fatalf("PartBytes = %d, want the image counted", PartBytes(shot))
	}
	stats := Tombstone(conv, Plan(conv, Policy{KeepTokens: 1}.Resolved()))
	stub := conv[2].Parts[0].(nacelle.ToolResult)
	if stats.Results != 1 || stats.Bytes < mediaBytes || len(stub.Images) != 0 {
		t.Errorf("stats = %+v, images left %d, want the image stubbed and counted as freed", stats, len(stub.Images))
	}
}

func TestPolicyOverIsStrict(t *testing.T) {
	policy := Policy{Ceiling: 100}
	if policy.Over(100) || !policy.Over(101) || (Policy{}).Over(1<<40) {
		t.Error("Over must be strict and never fire without a trigger")
	}
}
