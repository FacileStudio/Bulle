package runner

import (
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
	g.edits.calls[callID] = pendingCall{name: name, input: string(input)}
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
	g.edits.revisions[callID] = revision{tool: target, input: merged, edited: len(edit.Args) > 0}
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

// ruling is the decision for an approved call: the owner's edit or redirect
// when they made one, with the note that tells the model what ran.
func (g *gate) ruling(callID, asked string) nacelle.Decision {
	g.mu.Lock()
	rev, ok := g.edits.revisions[callID]
	delete(g.edits.revisions, callID)
	g.mu.Unlock()
	if !ok {
		return nacelle.Decision{}
	}
	return nacelle.Decision{Name: rev.tool, Input: rev.input, Note: revisionNote(asked, rev)}
}

func revisionNote(asked string, rev revision) string {
	var notes []string
	if rev.tool != asked {
		notes = append(notes, "The user chose "+rev.tool+" instead of "+asked+".")
	}
	if rev.edited {
		notes = append(notes, "The user edited the call before approving it; what ran was: "+string(rev.input)+".")
	}
	return strings.Join(notes, " ")
}
