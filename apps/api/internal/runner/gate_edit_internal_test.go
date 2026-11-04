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
	"github.com/FacileStudio/nacelle"
)

type echoTool struct{ name string }

func (e echoTool) Name() string           { return e.name }
func (e echoTool) Description() string    { return e.name }
func (e echoTool) Schema() map[string]any { return map[string]any{"type": "object"} }
func (e echoTool) Run(_ context.Context, input json.RawMessage) (string, error) {
	return e.name + " ran " + string(input), nil
}

func mailGate(t *testing.T) (*gate, map[string]nacelle.Tool) {
	t.Helper()
	st := store.NewMemory()
	em := &emitter{store: st, hub: newHub(), logger: slog.Default(), runID: "r1"}
	g := newGate(map[string]bool{mailSendTool: true}, st, em, "r1", time.Now)
	tools := map[string]nacelle.Tool{}
	for _, tool := range g.revisableTools([]nacelle.Tool{echoTool{mailSendTool}, echoTool{mailDraftTool}}) {
		tools[tool.Name()] = tool
	}
	return g, tools
}

func approveMail(t *testing.T, g *gate, callID, input string, edit Edit) (bool, error) {
	t.Helper()
	g.note(callID, mailSendTool, input)
	result := make(chan bool, 1)
	go func() { result <- g.approve(context.Background(), mailSendTool, json.RawMessage(input)) }()
	for deadline := time.Now().Add(2 * time.Second); !g.pending(callID); {
		if time.Now().After(deadline) {
			t.Fatalf("%s never asked the owner", callID)
		}
		time.Sleep(5 * time.Millisecond)
	}
	if err := g.decide(context.Background(), callID, true, "1", edit); err != nil {
		_ = g.decide(context.Background(), callID, false, "1", Edit{})
		<-result
		return false, err
	}
	return <-result, nil
}

func TestApprovedEditReplacesTheMail(t *testing.T) {
	g, tools := mailGate(t)
	input := `{"reply_to":42,"body":"Bonjour"}`
	ok, err := approveMail(t, g, "c1", input, Edit{Args: json.RawMessage(`{"body":"Bonjour Madame Nay","subject":"Visio"}`)})
	if err != nil || !ok {
		t.Fatalf("approve: %v %v", ok, err)
	}
	out, _ := tools[mailSendTool].Run(context.Background(), json.RawMessage(input))
	if !strings.Contains(out, `"body":"Bonjour Madame Nay"`) || !strings.Contains(out, `"reply_to":42`) || !strings.Contains(out, "edited") {
		t.Fatalf("the edited mail must run with the thread kept: %s", out)
	}
	again, _ := tools[mailSendTool].Run(context.Background(), json.RawMessage(input))
	if strings.Contains(again, "Madame") {
		t.Fatalf("an edit applies once: %s", again)
	}
}

func TestApprovedAsDraftSavesInstead(t *testing.T) {
	g, tools := mailGate(t)
	input := `{"to":["a@x"],"body":"hi"}`
	if ok, err := approveMail(t, g, "c2", input, Edit{As: "draft"}); err != nil || !ok {
		t.Fatalf("approve: %v %v", ok, err)
	}
	out, _ := tools[mailSendTool].Run(context.Background(), json.RawMessage(input))
	if !strings.HasPrefix(out, "The user chose courrier_save_draft") || !strings.Contains(out, "courrier_save_draft ran") {
		t.Fatalf("want the draft tool to run: %s", out)
	}
}

func TestEditRefusesFieldsOffTheAllowList(t *testing.T) {
	g, _ := mailGate(t)
	_, err := approveMail(t, g, "c3", `{"body":"hi"}`, Edit{Args: json.RawMessage(`{"account":99}`)})
	if !errors.Is(err, ErrBadEdit) {
		t.Fatalf("changing the mailbox must be refused, got %v", err)
	}
	_, err = approveMail(t, g, "c4", `{"body":"hi"}`, Edit{As: "delete"})
	if !errors.Is(err, ErrNotEditable) {
		t.Fatalf("an unknown target must be refused, got %v", err)
	}
}
