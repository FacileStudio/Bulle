package engine

import (
	"context"
	"iter"
	"testing"

	"github.com/FacileStudio/Bulle/core/nacelletest"
	"github.com/FacileStudio/nacelle"
)

type failing struct{ err error }

func (f failing) Name() string                       { return "failing" }
func (f failing) Capabilities() nacelle.Capabilities { return nacelle.Capabilities{} }
func (f failing) CountTokens(context.Context, nacelle.Request) (int64, error) {
	return 0, nil
}

func (f failing) Stream(context.Context, nacelle.Request) iter.Seq2[nacelle.Event, error] {
	return func(yield func(nacelle.Event, error) bool) {
		if !yield(nacelle.Event{Kind: nacelle.KindText, Text: "half an ans"}, nil) {
			return
		}
		yield(nacelle.Event{}, f.err)
	}
}

func newTestAgent(t *testing.T, b nacelle.Backend, tools ...nacelle.Tool) *nacelle.Agent {
	t.Helper()
	agent, err := nacelle.New(nacelle.Config{Backend: b, System: "test", Tools: tools})
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}
	return agent
}

func scripted(t *testing.T, turns []nacelletest.Turn, tools ...nacelle.Tool) *nacelle.Agent {
	t.Helper()
	return newTestAgent(t, nacelle.NewBackend(&nacelletest.Script{Turns: turns}), tools...)
}

func collectEvents(ch <-chan Event) []Event {
	var events []Event
	for ev := range ch {
		events = append(events, ev)
	}
	return events
}
