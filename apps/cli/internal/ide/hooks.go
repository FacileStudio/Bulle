package ide

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/FacileStudio/nacelle"

	"github.com/FacileStudio/Bulle/apps/cli/internal/diff"
)

// publisher is the nacelle hook this package registers. It holds the one
// piece of state that spans a call's two hooks: the change it is about to
// make, keyed by the call's id.
type publisher struct {
	srv     *Server
	mu      sync.Mutex
	pending map[string]diff.EditChange
}

// Hooks returns the publisher as the set of session hooks a run merges in
// with its own. It watches the two points a tool call is visible at and
// nothing else, so a hook that gates a call still runs first. A nil server
// publishes to nobody and adds no hook.
func (s *Server) Hooks() map[nacelle.HookPoint][]nacelle.Hook {
	if s == nil {
		return nil
	}
	p := &publisher{srv: s, pending: map[string]diff.EditChange{}}
	return map[nacelle.HookPoint][]nacelle.Hook{
		nacelle.BeforeToolCall: {p.before},
		nacelle.AfterToolCall:  {p.after},
	}
}

// before reports a call as started, and remembers the change it is about to
// make: a write_file's before side is only on disk until the call has run, so
// afterwards there is nothing left to read it from.
//
// A nil server is a session with the surface off. Its hooks stay installed and
// do nothing, which is what keeps the no-editor path from being a special case
// at every call site that merges them in.
func (p *publisher) before(_ context.Context, ev nacelle.HookEvent) nacelle.HookResult {
	if p.srv == nil {
		return nacelle.HookResult{}
	}
	id, input := callID(ev.Call), string(ev.Call.Input)
	if change, ok := diff.CaptureEdit(p.srv.opts.Root, ev.Call.Name, input); ok {
		p.remember(id, change)
	}
	p.srv.Publish(ToolStart(id, ev.Call.Name, pathOf(input)))
	return nacelle.HookResult{}
}

// after reports the call as finished, and the file it changed with it. A call
// that failed changed nothing worth marking, so it publishes no edit.
func (p *publisher) after(_ context.Context, ev nacelle.HookEvent) nacelle.HookResult {
	if p.srv == nil {
		return nacelle.HookResult{}
	}
	id := callID(ev.Call)
	path := pathOf(string(ev.Call.Input))
	p.srv.Publish(ToolDone(id, ev.Call.Name, ev.Err == nil, path))
	change, ok := p.take(id)
	if !ok || ev.Err != nil {
		return nacelle.HookResult{}
	}
	if marked, ok := p.change(ev.Call.Name, path, change); ok {
		p.srv.Publish(Edit(id, marked))
	}
	return nacelle.HookResult{}
}

// pathOf reads the file path out of a tool call's input. The protocol carries
// a path only where it means something, so every other call reports none.
func pathOf(input string) string {
	var fields struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(input), &fields); err != nil {
		return ""
	}
	return fields.Path
}
