package runner

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

func TestAskUserOnlyInWebChatsAndRejectsBadCards(t *testing.T) {
	if tools, _ := askUserTools(store.ChannelMatrix); len(tools) != 0 {
		t.Fatal("a Matrix run got ask_user")
	}
	tools, err := askUserTools(store.ChannelWeb)
	if err != nil || len(tools) != 1 {
		t.Fatalf("web run: %v %v", tools, err)
	}
	cases := map[string]bool{
		`{"questions":[{"question":"Which day?","options":[{"label":"Monday"},{"label":"Tuesday"}]}]}`: true,
		`{"questions":[{"question":"Which day?","options":[{"label":"Monday"}]}]}`:                     false,
		`{"questions":[{"question":"Which day?","options":[{"label":"Monday"},{"label":"Other"}]}]}`:   false,
		`{"questions":[{"question":"Which day?","options":[{"label":"Monday"},{"label":"monday"}]}]}`:  false,
		`{"questions":[]}`: false,
	}
	for input, ok := range cases {
		sink := &nacelle.ToolSink{}
		out, err := nacelle.RunTool(context.Background(), tools[0], nacelle.Invocation{ID: "c1", Name: "ask_user"}, json.RawMessage(input), sink)
		if (err == nil) != ok || sink.Yielded() != ok || (ok && !strings.Contains(out, "on a card")) {
			t.Errorf("%s: out %q err %v yielded %v", input, out, err, sink.Yielded())
		}
	}
}
