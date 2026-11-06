package runner

import (
	"context"
	"encoding/json"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
)

// suitePolicy describes one suite write for the approval card and hands the
// gate the scope an approval may be remembered on. The record carries the
// resource, the change and the session's space so the owner can judge the
// call without reading the model's raw JSON. A tool that never takes a grant
// keeps its description but no scope, so it is asked every time.
type suitePolicy struct {
	name     string
	defaults suite.Defaults
	space    store.Space
	gate     *gate
}

func (p suitePolicy) describe(_ context.Context, input json.RawMessage) (map[string]any, string, error) {
	scope, summary, err := suite.ScopeFor(p.name, input, p.defaults)
	if err != nil {
		return map[string]any{"tool": p.name, "args": string(input)}, "", nil
	}
	if p.space.Name != "" {
		summary["space"] = p.space.Name
	}
	if !p.gate.takesGrant(p.name) {
		return summary, "", nil
	}
	return summary, scope, nil
}

func (p suitePolicy) denied(context.Context, json.RawMessage) {}
