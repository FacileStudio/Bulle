package engine

import (
	"context"
	"errors"
	"strings"

	"github.com/FacileStudio/nacelle"
)

// Submit submits a user prompt and begins streaming events for the turn.
func (s *Session) Submit(ctx context.Context, prompt string) (<-chan Event, error) {
	if s.agent == nil {
		return nil, errors.New("engine: agent is required")
	}
	if prompt != "" {
		s.messages = Repair(append(s.messages, nacelle.UserText(prompt)))
	}
	if len(s.messages) == 0 {
		return nil, errors.New("engine: prompt is required for empty conversation")
	}
	events := make(chan Event, 64)
	go s.runStream(ctx, events)
	return events, nil
}

func (s *Session) runStream(ctx context.Context, events chan<- Event) {
	defer close(events)
	var pending strings.Builder
	defer func() { s.keep(pending.String()) }()
	for event, err := range s.agent.Stream(ctx, s.messages) {
		if err != nil {
			emit(ctx, events, NewError(err))
			return
		}
		switch event.Kind {
		case nacelle.KindText:
			pending.WriteString(event.Text)
		case nacelle.KindMessage:
			s.messages = append(s.messages, *event.Message)
			pending.Reset()
		}
		if !forward(ctx, event, events) {
			return
		}
	}
}

// forward hands the consumer the engine event a nacelle event stands for.
func forward(ctx context.Context, event nacelle.Event, events chan<- Event) bool {
	switch event.Kind {
	case nacelle.KindText:
		return emit(ctx, events, NewTextDelta(event.Text))
	case nacelle.KindThinking:
		return emit(ctx, events, NewThinking(event.Text))
	case nacelle.KindToolCall:
		if event.Tool != nil {
			return emit(ctx, events, NewToolCall(event.Tool))
		}
	case nacelle.KindToolResult:
		if event.Tool != nil && !event.Tool.Discarded {
			return emit(ctx, events, NewToolResult(event.Tool))
		}
	case nacelle.KindDone:
		return emit(ctx, events, NewTurnDone(event.Usage, event.Stop))
	}
	return true
}

func emit(ctx context.Context, events chan<- Event, ev Event) bool {
	select {
	case events <- ev:
		return true
	case <-ctx.Done():
		return false
	}
}
