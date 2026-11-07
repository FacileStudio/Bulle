package agent

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/FacileStudio/nacelle"

	"github.com/FacileStudio/Bulle/core/settings"
)

func TestADelegateWithNoApprovalGateMayCall(t *testing.T) {
	policy := delegateApprovals(nil)
	if policy == nil {
		t.Fatal("no policy was built, which the SDK reads as deny-all")
	}
	if policy(context.Background(), nacelle.Call{Name: "read_file"}).Deny != "" {
		t.Error("a call was refused in a session that asks about nothing")
	}
}

func TestADelegateStillObeysTheParentsGate(t *testing.T) {
	asked := ""
	gate := nacelle.ApproveBool(func(_ context.Context, name string, _ json.RawMessage) bool {
		asked = name
		return name == "read_file"
	})

	policy := delegateApprovals(gate)
	if policy(context.Background(), nacelle.Call{Name: "read_file"}).Deny != "" {
		t.Error("the gate allowed read_file and the delegate was refused anyway")
	}
	if policy(context.Background(), nacelle.Call{Name: "bash"}).Deny == "" {
		t.Error("the gate refused bash and the delegate ran it")
	}
	if asked != "bash" {
		t.Errorf("the gate was asked about %q, want the delegate's own calls", asked)
	}
}

// The delegate set is the parallel tool and its cancel tool: parallel_agents
// defaults on, so the mount is active in every ordinary session, and the cancel
// tool rides along so a model can stop a fan-out it started and no longer wants.
func TestParallelAgentsMountsTheParallelAndCancelTools(t *testing.T) {
	config := settings.Defaults("")
	on := true
	config.ParallelAgents = &on

	tools, err := withParallelAgents(config, &answeringStub{}, make([]nacelle.Tool, 0), nil)
	if err != nil {
		t.Fatalf("withParallelAgents: %v", err)
	}
	names := map[string]bool{}
	for _, tool := range tools {
		names[tool.Name()] = true
	}
	if len(tools) != 2 {
		t.Errorf("mounted %d tools, want the parallel tool and its cancel", len(tools))
	}
	if !names[nacelle.ParallelAgentsToolName] || !names[nacelle.ParallelCancelToolName] {
		t.Errorf("mounted %v, want parallel_agents and parallel_cancel", names)
	}
}

func TestResolveConcurrency(t *testing.T) {
	c16, c8, c4, c0 := 16, 8, 4, 0
	if got := resolveConcurrency(nil, nil); got != 16 {
		t.Errorf("got %d, want 16", got)
	}
	if got := resolveConcurrency(&c16, nil); got != 16 {
		t.Errorf("got %d, want 16", got)
	}
	if got := resolveConcurrency(&c8, nil); got != 8 {
		t.Errorf("got %d, want 8", got)
	}
	if got := resolveConcurrency(nil, &c4); got != 4 {
		t.Errorf("got %d, want 4", got)
	}
	if got := resolveConcurrency(&c16, &c4); got != 4 {
		t.Errorf("got %d, want 4", got)
	}
	if got := resolveConcurrency(&c8, &c4); got != 8 {
		t.Errorf("got %d, want 8", got)
	}
	if got := resolveConcurrency(&c0, &c0); got != 16 {
		t.Errorf("got %d, want 16", got)
	}
}
