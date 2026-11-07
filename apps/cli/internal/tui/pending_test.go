package tui

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/FacileStudio/Bulle/apps/cli/internal/approval"
	"github.com/FacileStudio/Bulle/apps/cli/internal/ask"
	"github.com/FacileStudio/Bulle/core/askuser"
)

// A main run that settles answers its own dead approval with a denial and
// leaves a detached subagent's live request in view, so the gate never stays
// locked on a request nobody will answer.
func TestSettleDeniesTheDeadApprovalAndKeepsTheChilds(t *testing.T) {
	m := sized()
	mainCtx, cancelMain := context.WithCancel(context.Background())
	childCtx, cancelChild := context.WithCancel(context.Background())
	defer cancelChild()
	m.run.busy = true
	m.run.cancel = cancelMain
	mainDecision := make(chan approval.Decision, 1)
	childDecision := make(chan approval.Decision, 1)
	m.parkApproval(approval.Request{Name: "main", Decision: mainDecision, Done: mainCtx.Done()})
	m.parkApproval(approval.Request{Name: "child", Decision: childDecision, Done: childCtx.Done()})

	m.settle()

	if d := <-mainDecision; d != approval.Denied {
		t.Errorf("dead approval got %v, want Denied", d)
	}
	if m.run.pending == nil || m.run.pending.Name != "child" {
		t.Fatalf("pending = %+v, want the child's request in view", m.run.pending)
	}
	m.key(tea.KeyPressMsg{Code: 'y', Text: "y"})
	if d := <-childDecision; d != approval.AllowedOnce {
		t.Errorf("child decision = %v, want AllowedOnce", d)
	}
}

// Requests queue oldest first: answering the shown one brings up the next.
func TestAnsweringAnApprovalShowsTheNextInLine(t *testing.T) {
	m := sized()
	first := make(chan approval.Decision, 1)
	second := make(chan approval.Decision, 1)
	m.parkApproval(approval.Request{Name: "first", Decision: first})
	m.parkApproval(approval.Request{Name: "second", Decision: second})

	m.key(tea.KeyPressMsg{Code: 'n', Text: "n"})

	if d := <-first; d != approval.Denied {
		t.Errorf("first = %v, want Denied", d)
	}
	if m.run.pending == nil || m.run.pending.Name != "second" {
		t.Fatalf("pending = %+v, want the second request", m.run.pending)
	}
}

// A question whose caller is gone is dismissed with nil on settle; a live one
// stays for the reader.
func TestSettleDismissesOnlyADeadQuestion(t *testing.T) {
	m := sized()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m.run.cancel = func() {}
	answers := make(chan []askuser.Answer, 1)
	q := []askuser.Question{{Question: "pick", Options: []askuser.Option{{Label: "a"}, {Label: "b"}}}}
	m.parkQuestion(ask.Request{Questions: q, Answers: answers, Done: ctx.Done()})

	m.settle()
	if m.run.question == nil {
		t.Fatal("a live question was dropped by the main run's settle")
	}
	cancel()
	m.settle()
	if m.run.question != nil {
		t.Fatal("a dead question survived settle")
	}
	if got := <-answers; got != nil {
		t.Errorf("dead question answered %v, want nil", got)
	}
}
