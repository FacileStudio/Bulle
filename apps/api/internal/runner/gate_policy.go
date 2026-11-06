package runner

import (
	"context"
	"encoding/json"
	"strings"
)

// approvalPolicy lets an outward tool shape its own approval: what the owner
// is shown in place of the model's arguments, how far one approval reaches
// (scope, empty for this call only), and what a denial has to undo.
type approvalPolicy interface {
	describe(ctx context.Context, input json.RawMessage) (record map[string]any, scope string, err error)
	denied(ctx context.Context, input json.RawMessage)
}

// describe is what the approval card shows for a call. A policy that cannot
// describe the call refuses it: the owner never approves something unseen.
func (g *gate) describe(ctx context.Context, name string, input json.RawMessage) (any, string, bool) {
	policy := g.policies[name]
	if policy == nil {
		return rawArgs(string(input)), "", true
	}
	record, scope, err := policy.describe(ctx, input)
	if err != nil {
		return nil, "", false
	}
	return record, scope, true
}

// granted reports whether an earlier approval in this run already covers the
// call: its scope, or the whole tool when the owner asked to remember it.
func (g *gate) granted(name, scope string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.grants[name+"\x00*"] && g.takesGrant(name) {
		return true
	}
	return scope != "" && g.grants[name+"\x00"+scope]
}

var ungrantable = []string{"_delete_", "_delete", "browser_", "web_fetch", "files_save"}

// takesGrant reports whether a remembered approval may cover name. Mail,
// deletions, file writes and every tool gated per URL stay asked each time:
// a grant keys on the tool alone, so it must not widen what an injected
// instruction can reach.
func (g *gate) takesGrant(name string) bool {
	if name == "" || name == mailSendTool || name == mailDraftTool || g.asks[name] != nil {
		return false
	}
	for _, part := range ungrantable {
		if strings.Contains(name, part) {
			return false
		}
	}
	_, mail := g.policies[name].(mailPolicy)
	return !mail
}

// settle records an approval for the rest of the run, or lets the policy
// undo what a denied call left behind.
func (g *gate) settle(ctx context.Context, name string, input json.RawMessage, scope string, approved bool) {
	if approved && scope != "" {
		g.mu.Lock()
		g.grants[name+"\x00"+scope] = true
		g.mu.Unlock()
		return
	}
	if policy := g.policies[name]; !approved && policy != nil {
		policy.denied(ctx, input)
	}
}
