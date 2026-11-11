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
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
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

func sendGate(t *testing.T) *gate {
	t.Helper()
	st := store.NewMemory()
	em := &emitter{store: st, hub: newHub(), logger: slog.Default(), runID: "r1"}
	return newGate(map[string]bool{"jobs_create": true}, st, em, "r1", time.Now)
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

func runGated(g *gate, callID, input string) chan ran {
	request := nacelle.Request{Tools: []nacelle.Tool{echoTool{"jobs_create"}}, Approve: g.approve}
	done := make(chan ran, 1)
	go func() {
		event, _, err := nacelletest.Call(context.Background(), request,
			nacelle.ToolCall{ID: callID, Name: "jobs_create", Input: json.RawMessage(input)})
		if err != nil {
			done <- ran{err: err}
			return
		}
		done <- ran{event.Result, event.Err}
	}()
	return done
}

func TestAnEditOnTheCardIsRefused(t *testing.T) {
	g := sendGate(t)
	done := runGated(g, "c1", `{"name":"brief"}`)
	waitPending(t, g, "c1")
	for _, edit := range []Edit{{Args: json.RawMessage(`{"name":"other"}`)}, {As: "draft"}} {
		if err := g.decide(context.Background(), "c1", true, "1", edit, false); !errors.Is(err, ErrNotEditable) {
			t.Fatalf("an edit must be refused, got %v", err)
		}
	}
	if err := g.decide(context.Background(), "c1", true, "1", Edit{}, false); err != nil {
		t.Fatal(err)
	}
	if got := <-done; got.err != nil || !strings.Contains(got.out, `jobs_create ran {"name":"brief"}`) {
		t.Fatalf("the call must run as the model sent it: %+v", got)
	}
}

func TestDeniedCallReadsAsRefused(t *testing.T) {
	g := sendGate(t)
	done := runGated(g, "c1", `{"name":"brief"}`)
	waitPending(t, g, "c1")
	if err := g.decide(context.Background(), "c1", false, "1", Edit{}, false); err != nil {
		t.Fatal(err)
	}
	if got := <-done; got.err == nil || !strings.Contains(got.err.Error(), `"jobs_create" was refused`) {
		t.Fatalf("want a refusal, got %+v", got)
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

func TestAnApprovalFillsOnlyAConfirmationForm(t *testing.T) {
	confirm := map[string]any{"type": "object", "properties": map[string]any{
		"confirm": map[string]any{"type": "boolean"}, "copy": map[string]any{"type": "string", "default": "me"},
	}, "required": []string{"confirm"}}
	content, err := suite.Acceptance(confirm)
	if err != nil || content["confirm"] != true || content["copy"] != "me" {
		t.Fatalf("content = %v, %v", content, err)
	}
	reason := map[string]any{"type": "object", "properties": map[string]any{"reason": map[string]any{"type": "string"}}, "required": []string{"reason"}}
	if _, err := suite.Acceptance(reason); err == nil {
		t.Fatal("a card must not invent an answer the owner never typed")
	}
}
