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
	record := suiteRecord(input, summary)
	if p.space.Name != "" {
		record["space"] = p.space.Name
	}
	if !p.gate.takesGrant(p.name) {
		return record, "", nil
	}
	return record, scope, nil
}

// suiteRecord is what the card shows: every argument the model passed, so the
// owner sees the change itself, then the resolved target over it. The bare id
// is dropped when the summary already names it by resource.
func suiteRecord(input json.RawMessage, summary map[string]any) map[string]any {
	record := map[string]any{}
	if err := json.Unmarshal(input, &record); err != nil {
		record = map[string]any{}
	}
	if resource, ok := summary["resource"].(string); ok && summary[resource] != nil {
		delete(record, "id")
	}
	for key, value := range summary {
		record[key] = value
	}
	return record
}

func (p suitePolicy) denied(context.Context, json.RawMessage) {}
