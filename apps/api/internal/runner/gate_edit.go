package runner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/FacileStudio/nacelle"
)

const (
	mailSendTool  = "courrier_send"
	mailDraftTool = "courrier_save_draft"
)

var (
	ErrNotEditable = errors.New("this call cannot be edited")
	ErrBadEdit     = errors.New("the edited call is not valid")
)

var editableFields = map[string][]string{mailSendTool: {"to", "cc", "bcc", "subject", "body"}}

var editTargets = map[string]map[string]string{mailSendTool: {"": mailSendTool, "send": mailSendTool, "draft": mailDraftTool}}

// Edit is what the owner changed on an approval card: which tool runs in the
// end (As, empty for the one the model asked for) and the fields they rewrote.
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
	calls     map[string]pendingCall
	revisions map[string]revision
}

type revision struct {
	tool   string
	input  json.RawMessage
	edited bool
}

func (g *gate) remember(callID, name string, input json.RawMessage) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.edits.calls == nil {
		g.edits.calls = map[string]pendingCall{}
	}
	g.edits.calls[callID] = pendingCall{name: name, input: normalize(string(input))}
}

// revise turns an approved edit into the call that runs in place of the one
// the model asked for. Only fields on the tool's allow-list can change.
func (g *gate) revise(callID string, edit Edit) error {
	g.mu.Lock()
	call, ok := g.edits.calls[callID]
	g.mu.Unlock()
	if !ok {
		return ErrNoPendingApproval
	}
	target, ok := editTargets[call.name][edit.As]
	if !ok {
		return ErrNotEditable
	}
	merged, err := mergeEdit(call.input, edit.Args, editableFields[call.name])
	if err != nil {
		return err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.edits.revisions == nil {
		g.edits.revisions = map[string]revision{}
	}
	g.edits.revisions[call.name+"\x00"+call.input] = revision{tool: target, input: merged, edited: len(edit.Args) > 0}
	return nil
}

func mergeEdit(original string, args json.RawMessage, allowed []string) (json.RawMessage, error) {
	var base map[string]any
	if err := json.Unmarshal([]byte(original), &base); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBadEdit, err)
	}
	if len(args) == 0 {
		return json.RawMessage(original), nil
	}
	var changes map[string]any
	if err := json.Unmarshal(args, &changes); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBadEdit, err)
	}
	for key, value := range changes {
		if !slices.Contains(allowed, key) {
			return nil, fmt.Errorf("%w: %s cannot be changed", ErrBadEdit, key)
		}
		base[key] = value
	}
	out, err := json.Marshal(base)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBadEdit, err)
	}
	return out, nil
}

func (g *gate) takeRevision(name string, input json.RawMessage) (revision, bool) {
	key := name + "\x00" + normalize(string(input))
	g.mu.Lock()
	defer g.mu.Unlock()
	rev, ok := g.edits.revisions[key]
	delete(g.edits.revisions, key)
	return rev, ok
}

// revisable runs an approved call as the owner left it on the card: with
// their edits, or as a different tool (a draft in place of a send).
type revisable struct {
	nacelle.Tool
	gate  *gate
	tools map[string]nacelle.Tool
}

func (t revisable) Run(ctx context.Context, input json.RawMessage) (string, error) {
	rev, ok := t.gate.takeRevision(t.Name(), input)
	if !ok {
		return t.Tool.Run(ctx, input)
	}
	tool := t.tools[rev.tool]
	if tool == nil {
		return "", fmt.Errorf("%s is not available", rev.tool)
	}
	out, err := tool.Run(ctx, rev.input)
	if err != nil {
		return "", err
	}
	return revisionNote(t.Name(), rev) + out, nil
}

func revisionNote(asked string, rev revision) string {
	var notes []string
	if rev.tool != asked {
		notes = append(notes, "The user chose "+rev.tool+" instead of "+asked+".")
	}
	if rev.edited {
		notes = append(notes, "The user edited the call before approving it; what ran was: "+string(rev.input)+".")
	}
	if len(notes) == 0 {
		return ""
	}
	return strings.Join(notes, " ") + "\n"
}

// revisableTools wraps every editable tool so an approved edit takes effect.
func (g *gate) revisableTools(tools []nacelle.Tool) []nacelle.Tool {
	byName := map[string]nacelle.Tool{}
	for _, tool := range tools {
		byName[tool.Name()] = tool
	}
	out := make([]nacelle.Tool, len(tools))
	for i, tool := range tools {
		if _, ok := editTargets[tool.Name()]; ok {
			out[i] = revisable{Tool: tool, gate: g, tools: byName}
			continue
		}
		out[i] = tool
	}
	return out
}
