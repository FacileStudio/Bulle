package runner

import (
	"encoding/json"
	"errors"
)

// ErrNotEditable answers an approval that tries to rewrite the call or run
// another tool. No tool takes edits since Courrier moved to MCP: its server
// owns the draft, so the user changes the mail there, not on the card.
var ErrNotEditable = errors.New("this call cannot be edited")

// Edit is what an older client sent to rewrite a call on its approval card:
// which tool runs in the end (As) and the fields it rewrote. Any non-empty
// edit is refused.
type Edit struct {
	As   string
	Args json.RawMessage
}

func (e Edit) empty() bool { return e.As == "" && len(e.Args) == 0 }

type pendingCall struct {
	name  string
	input string
}

type editState struct {
	calls map[string]pendingCall
}

func (g *gate) remember(callID, name string, input json.RawMessage) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.edits.calls == nil {
		g.edits.calls = map[string]pendingCall{}
	}
	g.edits.calls[callID] = pendingCall{name: name, input: string(input)}
}
