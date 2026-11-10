package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
)

// RunWaiting is the status a run view reports while a call waits for the
// owner's approval. The stored run stays running meanwhile.
const RunWaiting = "waiting"

const summaryLimit = 600

// RunView is a run as a caller without the chat UI sees it: its status, the
// last thing the agent said, the objects its calls changed, and the calls
// that wait for the owner's approval.
type RunView struct {
	ID         string
	SessionID  string
	Status     string
	Summary    string
	Outputs    []RunOutput
	Pending    []PendingApproval
	StartedAt  time.Time
	FinishedAt *time.Time
}

// RunOutput is one call of the run that changed the user's data and
// succeeded. ID and URL are read from the tool's answer when it names them.
type RunOutput struct {
	CallID  string
	Tool    string
	App     string
	ID      string
	URL     string
	Preview string
}

// PendingApproval is a call held for the owner. It is answered through the
// same approval endpoint the chat's card uses.
type PendingApproval struct {
	CallID    string
	Tool      string
	Args      json.RawMessage
	Grantable bool
}

// ViewRun reads one of the user's runs from its stored events and, while it
// is live, from its approval gate.
func (r *Runner) ViewRun(ctx context.Context, userID int64, runID string) (RunView, error) {
	run, err := r.cfg.Store.GetRun(ctx, userID, runID)
	if err != nil {
		return RunView{}, err
	}
	events, err := r.cfg.Store.Events(ctx, runID, 0)
	if err != nil {
		return RunView{}, err
	}
	view := RunView{ID: run.ID, SessionID: run.SessionID, Status: run.Status, StartedAt: run.StartedAt, FinishedAt: run.FinishedAt}
	reader := runReader{changes: r.changingTools(), waiting: r.waitingCalls(runID), starts: map[string]startedCall{}}
	for _, event := range events {
		reader.read(event)
	}
	view.Summary, view.Outputs, view.Pending = clipSummary(reader.text.String()), reader.outputs, reader.pending
	if view.Status == store.RunRunning && len(view.Pending) > 0 {
		view.Status = RunWaiting
	}
	return view, nil
}

func (r *Runner) changingTools() map[string]bool {
	changes := map[string]bool{}
	for name := range r.outward {
		changes[name] = true
	}
	for _, name := range append(suite.SelfWrites(), suite.InboxTools()...) {
		changes[name] = true
	}
	return changes
}

func (r *Runner) waitingCalls(runID string) map[string]bool {
	r.mu.Lock()
	lr := r.live[runID]
	r.mu.Unlock()
	waiting := map[string]bool{}
	if lr == nil {
		return waiting
	}
	lr.gate.mu.Lock()
	defer lr.gate.mu.Unlock()
	for id := range lr.gate.waiting {
		waiting[id] = true
	}
	return waiting
}

type startedCall struct {
	name string
}

type runReader struct {
	changes map[string]bool
	waiting map[string]bool
	starts  map[string]startedCall
	text    strings.Builder
	outputs []RunOutput
	pending []PendingApproval
}

func (rr *runReader) read(event store.Event) {
	var payload map[string]json.RawMessage
	if json.Unmarshal(event.Payload, &payload) != nil {
		return
	}
	switch event.Type {
	case EventText:
		rr.text.WriteString(field(payload, "delta"))
	case EventToolStart:
		rr.text.Reset()
		rr.starts[field(payload, "call_id")] = startedCall{name: field(payload, "name")}
	case EventToolEnd:
		rr.ended(payload)
	case EventInterrupt:
		rr.interrupted(payload)
	}
}

func (rr *runReader) ended(payload map[string]json.RawMessage) {
	id := field(payload, "call_id")
	call, ok := rr.starts[id]
	var succeeded bool
	_ = json.Unmarshal(payload["ok"], &succeeded)
	if !ok || !succeeded || !rr.changes[call.name] {
		return
	}
	preview := field(payload, "output_preview")
	objectID, url := objectOf(preview)
	rr.outputs = append(rr.outputs, RunOutput{
		CallID: id, Tool: call.name, App: appOfTool(call.name), ID: objectID, URL: url, Preview: clipSummary(preview),
	})
}

func (rr *runReader) interrupted(payload map[string]json.RawMessage) {
	id := field(payload, "call_id")
	if !rr.waiting[id] {
		return
	}
	var grantable bool
	_ = json.Unmarshal(payload["grantable"], &grantable)
	args := payload["args"]
	if len(args) == 0 {
		args = json.RawMessage("null")
	}
	rr.pending = append(rr.pending, PendingApproval{CallID: id, Tool: field(payload, "tool"), Args: args, Grantable: grantable})
}

func field(payload map[string]json.RawMessage, key string) string {
	var value string
	_ = json.Unmarshal(payload[key], &value)
	return value
}

// objectOf reads the id and url a tool's JSON answer names at its top level.
func objectOf(output string) (string, string) {
	var object map[string]any
	if json.Unmarshal([]byte(output), &object) != nil {
		return "", ""
	}
	url, _ := object["url"].(string)
	switch id := object["id"].(type) {
	case string:
		return id, url
	case float64:
		return fmt.Sprintf("%.0f", id), url
	}
	return "", url
}
