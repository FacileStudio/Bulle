package runner

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/core/nacelletest"
	"github.com/FacileStudio/nacelle"
)

type echoTool struct{ name string }

func (e echoTool) Spec() nacelle.Spec {
	return nacelle.Spec{Name: e.name, Description: e.name, Schema: map[string]any{"type": "object"}}
}

func (e echoTool) Run(_ context.Context, call nacelle.Call, _ func(string)) (nacelle.Result, error) {
	return nacelle.Result{Text: e.name + " ran " + string(call.Input)}, nil
}

func mailGate(t *testing.T) *gate {
	t.Helper()
	st := store.NewMemory()
	em := &emitter{store: st, hub: newHub(), logger: slog.Default(), runID: "r1"}
	return newGate(map[string]bool{mailSendTool: true}, st, em, "r1", time.Now)
}

func (g *gate) allows(id, name, input string) bool {
	return g.approve(context.Background(), nacelle.Call{ID: id, Name: name, Input: json.RawMessage(input)}).Deny == ""
}

func waitPending(t *testing.T, g *gate, ids ...string) {
	t.Helper()
	for _, id := range ids {
		for deadline := time.Now().Add(2 * time.Second); !g.pending(id); {
			if time.Now().After(deadline) {
				t.Fatalf("%s never asked the owner", id)
			}
			time.Sleep(5 * time.Millisecond)
		}
	}
}

type ran struct {
	out string
	err error
}

func runMail(g *gate, callID, input string) chan ran {
	tools := []nacelle.Tool{echoTool{mailSendTool}, echoTool{mailDraftTool}}
	request := nacelle.Request{Tools: tools, Approve: g.approve}
	done := make(chan ran, 1)
	go func() {
		event, _, err := nacelletest.Call(context.Background(), request,
			nacelle.ToolCall{ID: callID, Name: mailSendTool, Input: json.RawMessage(input)})
		if err != nil {
			done <- ran{err: err}
			return
		}
		done <- ran{event.Result, event.Err}
	}()
	return done
}

func approveMail(t *testing.T, g *gate, callID, input string, edit Edit) (string, error) {
	t.Helper()
	done := runMail(g, callID, input)
	waitPending(t, g, callID)
	if err := g.decide(context.Background(), callID, true, "1", edit, false); err != nil {
		_ = g.decide(context.Background(), callID, false, "1", Edit{}, false)
		<-done
		return "", err
	}
	got := <-done
	return got.out, got.err
}

func TestApprovedEditReplacesTheMail(t *testing.T) {
	g := mailGate(t)
	out, err := approveMail(t, g, "c1", `{"reply_to":42,"body":"Bonjour"}`, Edit{Args: json.RawMessage(`{"body":"Bonjour Madame Nay","subject":"Visio"}`)})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `courrier_send ran {"body":"Bonjour Madame Nay","reply_to":42,"subject":"Visio"}`) || !strings.HasPrefix(out, "The user edited the call") {
		t.Fatalf("the edited mail must run with the thread kept: %s", out)
	}
}

func TestApprovedAsDraftSavesInstead(t *testing.T) {
	g := mailGate(t)
	out, err := approveMail(t, g, "c2", `{"to":["a@x"],"body":"hi"}`, Edit{As: "draft"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, "The user chose courrier_save_draft") || !strings.Contains(out, "courrier_save_draft ran") {
		t.Fatalf("want the draft tool to run: %s", out)
	}
}

func TestTwinCallsGetTheirOwnEdit(t *testing.T) {
	g := mailGate(t)
	input := `{"to":["a@x"],"body":"hi"}`
	first, second := runMail(g, "c1", input), runMail(g, "c2", input)
	waitPending(t, g, "c1", "c2")
	if err := g.decide(context.Background(), "c2", true, "1", Edit{}, false); err != nil {
		t.Fatal(err)
	}
	if err := g.decide(context.Background(), "c1", true, "1", Edit{Args: json.RawMessage(`{"body":"edited"}`)}, false); err != nil {
		t.Fatal(err)
	}
	edited, plain := <-first, <-second
	if !strings.Contains(edited.out, `"body":"edited"`) || strings.Contains(plain.out, "edited") {
		t.Fatalf("the edit must run on the call it was made on: %q / %q", edited.out, plain.out)
	}
}

func TestDeniedMailReadsAsRefused(t *testing.T) {
	g := mailGate(t)
	done := runMail(g, "c1", `{"body":"hi"}`)
	waitPending(t, g, "c1")
	if err := g.decide(context.Background(), "c1", false, "1", Edit{}, false); err != nil {
		t.Fatal(err)
	}
	if got := <-done; got.err == nil || !strings.Contains(got.err.Error(), `"courrier_send" was refused`) {
		t.Fatalf("want a refusal, got %+v", got)
	}
}

func TestEditRefusesFieldsOffTheAllowList(t *testing.T) {
	g := mailGate(t)
	_, err := approveMail(t, g, "c3", `{"body":"hi"}`, Edit{Args: json.RawMessage(`{"account":99}`)})
	if !errors.Is(err, ErrBadEdit) {
		t.Fatalf("changing the mailbox must be refused, got %v", err)
	}
	_, err = approveMail(t, g, "c4", `{"body":"hi"}`, Edit{As: "delete"})
	if !errors.Is(err, ErrNotEditable) {
		t.Fatalf("an unknown target must be refused, got %v", err)
	}
}

type failingApprovals struct{ store.Store }

func (failingApprovals) CreateApproval(context.Context, string, string) error {
	return errors.New("database is gone")
}

func TestUnrecordedApprovalRefusesAndSaysSo(t *testing.T) {
	st := failingApprovals{store.NewMemory()}
	em := &emitter{store: st, hub: newHub(), logger: slog.Default(), runID: "r1"}
	g := newGate(map[string]bool{"send": true}, st, em, "r1", time.Now)
	if g.allows("c1", "send", `{}`) {
		t.Fatal("a call whose approval was not recorded ran")
	}
	events, _ := st.Events(context.Background(), "r1", 0)
	if len(events) != 1 || events[0].Type != EventApprovalError || !strings.Contains(string(events[0].Payload), `"call_id":"c1"`) {
		t.Fatalf("want one APPROVAL_ERROR event, got %+v", events)
	}
}
