package runner

import (
	"context"
	"encoding/json"
	"strconv"
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
// call: the same tool, on the same scope. An empty scope never covers, so a
// target the policy could not name is asked every time.
func (g *gate) granted(name, scope string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return scope != "" && g.grants[grantKey(name, scope)]
}

// grantKey is where a grant on scope is kept: per tool, except the link
// grant, which covers sonar_fetch and browser_navigate together.
func grantKey(name, scope string) string {
	if scope == linkScope {
		return linkScope
	}
	return name + "\x00" + scope
}

// grantable reports whether the owner may approve a call for the rest of the
// run: a scoped call of a tool that takes grants, or the link card.
func (g *gate) grantable(name, scope string) bool {
	return scope == linkScope || (scope != "" && g.takesGrant(name))
}

// linkCard is what the owner sees when a tainted run wants to follow one more
// link it found in a tool's result.
func linkCard(input json.RawMessage) map[string]any {
	var in struct {
		URL string `json:"url"`
	}
	_ = json.Unmarshal(input, &in)
	return map[string]any{
		"url": in.URL,
		"why": "This run already followed " + strconv.Itoa(linkBudget) + " links found in mail, pages or files. " +
			"Which links a run picks can leak what it read, so each further one needs your approval.",
	}
}

var ungrantable = []string{"_delete_", "_delete", "browser_"}

// takesGrant reports whether a remembered approval may cover name. Deletions
// and browser requests stay asked each time: a grant keys on the
// tool alone, so it must not widen what an injected instruction can reach.
// A server's confirmation never takes one either: it has no scope.
func (g *gate) takesGrant(name string) bool {
	if name == "" {
		return false
	}
	for _, part := range ungrantable {
		if strings.Contains(name, part) {
			return false
		}
	}
	return true
}

func (g *gate) addPolicy(name string, policy approvalPolicy) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.policies == nil {
		g.policies = map[string]approvalPolicy{}
	}
	g.policies[name] = policy
}

// settle records an approval for the rest of the run, or lets the policy
// undo what a denied call left behind. Approving one link covers that call
// only: the owner grants more links on the card itself.
func (g *gate) settle(ctx context.Context, name string, input json.RawMessage, scope string, approved bool) {
	if approved && scope == linkScope {
		return
	}
	if approved && scope != "" {
		g.mu.Lock()
		g.grants[grantKey(name, scope)] = true
		g.mu.Unlock()
		return
	}
	if policy := g.policies[name]; !approved && policy != nil {
		policy.denied(ctx, input)
	}
}
