package agent

import (
	"context"
	"encoding/json"
	"math"

	"github.com/FacileStudio/nacelle"
)

const callTool = "call_tool"

// shapeCalls puts two rules in front of the session's approval gate, so they
// hold for the main run and every delegate alike.
//
// A call_tool is asked under the bridged tool it names. In catalog mode every
// bridged tool runs through that one tool, so asked as-is the gate would show
// "call_tool" for all of them and one Allow-for-session would open them all.
//
// A bash call has the co-author trailer added to any git commit in it, when
// coauthor is on, and the gate sees the command as it will run.
func shapeCalls(approve nacelle.Approve, coauthor bool) nacelle.Approve {
	if approve == nil {
		approve = func(context.Context, nacelle.Call) nacelle.Decision { return nacelle.Decision{} }
	}
	return func(ctx context.Context, call nacelle.Call) nacelle.Decision {
		switch call.Name {
		case callTool:
			return nacelle.Decision{Deny: approve(ctx, bridged(call)).Deny}
		case shellTool:
			call.Input = coAuthorInput(call.Input, coauthor)
			decision := approve(ctx, call)
			if decision.Input == nil {
				decision.Input = call.Input
			}
			return decision
		}
		return approve(ctx, call)
	}
}

// bridged is a call_tool call as the bridged tool it names. Input without a
// legible name stands as it is, malformed input being the refusal the gate
// exists to catch.
func bridged(call nacelle.Call) nacelle.Call {
	var in struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(call.Input, &in); err != nil || in.Name == "" {
		return call
	}
	call.Name, call.Input = in.Name, normalizeCallToolArgs(in.Arguments)
	return call
}

// UnmarshalJSON accepts timeout, in seconds, as the pre-v0.32 spelling of
// timeout_ms. timeout_ms wins when both are sent.
func (in *shellCommandInput) UnmarshalJSON(data []byte) error {
	type plain shellCommandInput
	var legacy struct {
		Timeout int64 `json:"timeout"`
	}
	if err := json.Unmarshal(data, (*plain)(in)); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &legacy); err != nil {
		return err
	}
	if in.TimeoutMs == 0 && legacy.Timeout > 0 {
		in.TimeoutMs = min(legacy.Timeout, math.MaxInt64/1000) * 1000
	}
	return nil
}
