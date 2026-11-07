// Package nacelletest runs nacelle tools the way an agent run does, for
// tests: one call outside the loop, or a scripted model driven through
// nacelle's own loop so Approve, hooks and routing all apply.
package nacelletest

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"sync"

	"github.com/FacileStudio/nacelle"
)

// Run calls tool outside the loop with input and answers the text it
// returned.
func Run(ctx context.Context, tool nacelle.Tool, input json.RawMessage) (string, error) {
	result, err := tool.Run(ctx, nacelle.Call{Name: tool.Spec().Name, Input: input}, nil)
	return result.Text, err
}

// Turn is one scripted model turn: the calls it asks for, or, with none, a
// text answer.
type Turn struct {
	Thinking string
	Text     string
	Calls    []nacelle.ToolCall
	Usage    nacelle.Usage
}

// Script is a Provider that plays its turns in order and answers an empty
// finished turn once they run out.
type Script struct {
	Turns []Turn

	mu       sync.Mutex
	next     int
	requests []nacelle.Request
}

// Name identifies the script.
func (s *Script) Name() string { return "script" }

// Capabilities reports thinking, so a request that asks for it is accepted.
func (s *Script) Capabilities() nacelle.Capabilities { return nacelle.Capabilities{Thinking: true} }

// CountTokens reports nothing.
func (s *Script) CountTokens(context.Context, nacelle.Request) (int64, error) { return 0, nil }

// Requests answers every request the loop sent, in order.
func (s *Script) Requests() []nacelle.Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]nacelle.Request(nil), s.requests...)
}

// Turn streams the next scripted turn.
func (s *Script) Turn(_ context.Context, request nacelle.Request) iter.Seq2[nacelle.Event, error] {
	s.mu.Lock()
	var turn Turn
	if s.next < len(s.Turns) {
		turn = s.Turns[s.next]
	}
	s.next++
	s.requests = append(s.requests, request)
	s.mu.Unlock()
	return func(yield func(nacelle.Event, error) bool) {
		for _, event := range turn.Events() {
			if !yield(event, nil) {
				return
			}
		}
	}
}

// Events is the turn as a provider streams it: reasoning and text deltas, a
// KindToolCall per call, the KindTurn and the assistant message.
func (t Turn) Events() []nacelle.Event {
	var events []nacelle.Event
	message := nacelle.Message{Role: nacelle.RoleAssistant}
	if t.Thinking != "" {
		events = append(events, nacelle.Event{Kind: nacelle.KindThinking, Text: t.Thinking})
		message.Parts = append(message.Parts, nacelle.Reasoning{Text: t.Thinking})
	}
	if t.Text != "" {
		events = append(events, nacelle.Event{Kind: nacelle.KindText, Text: t.Text})
		message.Parts = append(message.Parts, nacelle.Text{Text: t.Text})
	}
	stop := nacelle.StopEnd
	for index, call := range t.Calls {
		call.Finished = true
		message.Parts = append(message.Parts, call)
		events = append(events, nacelle.Event{Kind: nacelle.KindToolCall, Tool: &nacelle.ToolEvent{
			ID: call.ID, Index: index, Name: call.Name, Input: string(call.Input),
		}})
		stop = nacelle.StopTools
	}
	return append(events,
		nacelle.Event{Kind: nacelle.KindTurn, Stop: stop, Usage: t.Usage},
		nacelle.Event{Kind: nacelle.KindMessage, Message: &message},
	)
}

// Call runs one call through the real loop under request's tools, Approve
// and hooks, and answers its KindToolResult with the Stop the run ended on.
func Call(ctx context.Context, request nacelle.Request, call nacelle.ToolCall) (*nacelle.ToolEvent, nacelle.Stop, error) {
	script := &Script{Turns: []Turn{{Calls: []nacelle.ToolCall{call}}}}
	var result *nacelle.ToolEvent
	for event, err := range nacelle.NewBackend(script).Stream(ctx, request) {
		if err != nil {
			return nil, "", err
		}
		switch event.Kind {
		case nacelle.KindToolResult:
			result = event.Tool
		case nacelle.KindDone:
			if result == nil {
				return nil, event.Stop, fmt.Errorf("nacelletest: %s produced no result", call.Name)
			}
			return result, event.Stop, nil
		}
	}
	return nil, "", fmt.Errorf("nacelletest: the run for %s never finished", call.Name)
}
