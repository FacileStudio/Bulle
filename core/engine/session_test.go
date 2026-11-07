package engine

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/core/nacelletest"
	"github.com/FacileStudio/nacelle"
)

func roles(messages []nacelle.Message) string {
	var out []string
	for _, m := range messages {
		out = append(out, string(m.Role))
	}
	return strings.Join(out, ",")
}

func TestSessionSubmitNilAgent(t *testing.T) {
	if _, err := NewSession(nil, nil).Submit(context.Background(), "hello"); err == nil {
		t.Fatalf("expected error with nil agent")
	}
}

func TestSessionSubmitEmptyPromptEmptyConversation(t *testing.T) {
	if _, err := NewSession(scripted(t, nil), nil).Submit(context.Background(), ""); err == nil {
		t.Fatalf("expected error with empty prompt on empty conversation")
	}
}

func TestSessionKeepsWhatTheRunCommitted(t *testing.T) {
	echo, err := nacelle.NewImageTool("look", "look", func(context.Context, struct{}) (string, []nacelle.Image, error) {
		return "a picture", []nacelle.Image{{MediaType: "image/png", Data: []byte("png")}}, nil
	}, nacelle.ToolOptions{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	agent := scripted(t, []nacelletest.Turn{
		{Thinking: "hm", Text: "looking", Calls: []nacelle.ToolCall{{ID: "c1", Name: "look", Input: json.RawMessage(`{}`)}}},
		{Text: "found it"},
	}, echo)
	s := NewSession(agent, nil)
	ch, err := s.Submit(context.Background(), "what is there?")
	if err != nil {
		t.Fatal(err)
	}
	var kinds []string
	for _, ev := range collectEvents(ch) {
		kinds = append(kinds, string(ev.Kind))
	}
	if strings.Join(kinds, ",") != "thinking,text_delta,tool_call,tool_result,text_delta,turn_done" {
		t.Fatalf("events = %v", kinds)
	}
	messages := s.Messages()
	if roles(messages) != "user,assistant,user,assistant" {
		t.Fatalf("roles = %s", roles(messages))
	}
	if _, ok := messages[1].Parts[0].(nacelle.Reasoning); !ok {
		t.Errorf("the reasoning was lost: %+v", messages[1])
	}
	if result, _ := messages[2].Parts[0].(nacelle.ToolResult); len(result.Images) != 1 {
		t.Errorf("the tool image was lost: %+v", messages[2])
	}
}

func TestSessionSubmitMultiTurn(t *testing.T) {
	s := NewSession(scripted(t, []nacelletest.Turn{{Text: "one"}, {Text: "two"}}), nil)
	ch1, _ := s.Submit(context.Background(), "turn 1")
	collectEvents(ch1)
	ch2, _ := s.Submit(context.Background(), "turn 2")
	collectEvents(ch2)
	if got := roles(s.Messages()); got != "user,assistant,user,assistant" {
		t.Fatalf("roles = %s", got)
	}
}

func TestSessionKeepsTheTextOfATurnCutShort(t *testing.T) {
	s := NewSession(newTestAgent(t, failing{err: errors.New("backend failure")}), nil)
	ch, _ := s.Submit(context.Background(), "fail")
	events := collectEvents(ch)
	if last := events[len(events)-1]; last.Kind != EventError {
		t.Fatalf("expected an error event last, got %v", events)
	}
	messages := s.Messages()
	if roles(messages) != "user,assistant" || messages[1].Parts[0].(nacelle.Text).Text != "half an ans" {
		t.Fatalf("messages = %+v, want the streamed text kept", messages)
	}
}

func TestSessionSubmitContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ch, _ := NewSession(scripted(t, []nacelletest.Turn{{Text: "stream"}}), nil).Submit(ctx, "cancel")
	collectEvents(ch)
}
