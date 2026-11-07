package agent

import (
	"context"
	"encoding/json"
	"math"

	"github.com/FacileStudio/nacelle"
)

// shapeCalls puts a rule in front of the session's approval gate, so it holds
// for the main run and every delegate alike: a bash call has the co-author
// trailer added to any git commit in it, when coauthor is on, and the gate
// sees the command as it will run.
func shapeCalls(approve nacelle.Approve, coauthor bool) nacelle.Approve {
	if approve == nil {
		approve = func(context.Context, nacelle.Call) nacelle.Decision { return nacelle.Decision{} }
	}
	return func(ctx context.Context, call nacelle.Call) nacelle.Decision {
		if call.Name != shellTool {
			return approve(ctx, call)
		}
		call.Input = coAuthorInput(call.Input, coauthor)
		decision := approve(ctx, call)
		if decision.Input == nil {
			decision.Input = call.Input
		}
		return decision
	}
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
