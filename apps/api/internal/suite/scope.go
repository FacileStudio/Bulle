package suite

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ScopeFor names the target one write tool call aims at, resolved through the
// run's Defaults when the call omits its workspace, account or calendar. The
// scope keys a remembered approval to that target and nothing wider. An error
// means the input could not be read; an empty scope with no error means no
// target was named. A caller must ask every time for both and must never
// treat either as grantable.
func ScopeFor(name string, input []byte, d Defaults) (string, map[string]any, error) {
	name = CanonicalTool(name)
	input = CanonicalInput(name, input)
	switch {
	case strings.HasPrefix(name, "opus_"):
		return scopeOpus(name, input, d)
	case strings.HasPrefix(name, "glouton_"):
		return scopeGlouton(name, input, d)
	case strings.HasPrefix(name, "sablier_"), strings.HasPrefix(name, "nuage_"):
		return scopeSablierFiles(name, input, d)
	}
	return "", nil, fmt.Errorf("suite: no scope rule for %q", name)
}

func scopeDecode[In any](name string, input []byte, d Defaults, fill func(In, Defaults) (string, map[string]any)) (string, map[string]any, error) {
	var in In
	if err := json.Unmarshal(input, &in); err != nil {
		return "", nil, fmt.Errorf("suite: %s: reading arguments: %w", name, err)
	}
	scope, summary := fill(in, d)
	return scope, summary, nil
}

func scopeValue[T string | int64](kind string, value T) string {
	var zero T
	if value == zero {
		return ""
	}
	return kind + ":" + fmt.Sprint(value)
}

func scopeJoin(base, part string) string {
	if base == "" || part == "" {
		return base
	}
	return base + "/" + part
}

func scopeOpus(name string, input []byte, d Defaults) (string, map[string]any, error) {
	switch name {
	case "opus_create_task":
		return scopeDecode(name, input, d, func(in opusCreateInput, d Defaults) (string, map[string]any) {
			workspace := d.workspace(in.Workspace)
			return scopeJoin(scopeValue("workspace", workspace), scopeValue("project", in.Project)), map[string]any{"resource": "task", "action": "create", "workspace": workspace, "project": in.Project, "title": in.Title}
		})
	case "opus_update_task":
		return scopeDecode(name, input, d, func(in opusUpdateInput, _ Defaults) (string, map[string]any) {
			return scopeValue("task", in.TaskID), map[string]any{"resource": "task", "action": "update", "task": in.TaskID}
		})
	case "opus_delete_task":
		return scopeDecode(name, input, d, func(in opusTaskInput, _ Defaults) (string, map[string]any) {
			return scopeValue("task", in.TaskID), map[string]any{"resource": "task", "action": "delete", "task": in.TaskID}
		})
	case "opus_create_project":
		return scopeDecode(name, input, d, func(in opusCreateProjectInput, d Defaults) (string, map[string]any) {
			workspace := d.workspace(in.Workspace)
			return scopeValue("workspace", workspace), map[string]any{"resource": "project", "action": "create", "workspace": workspace, "name": in.Name}
		})
	case "opus_update_project":
		return scopeDecode(name, input, d, func(in opusUpdateProjectInput, _ Defaults) (string, map[string]any) {
			return scopeValue("project", in.ProjectID), map[string]any{"resource": "project", "action": "update", "project": in.ProjectID}
		})
	case "opus_delete_project":
		return scopeDecode(name, input, d, func(in opusProjectInput, _ Defaults) (string, map[string]any) {
			return scopeValue("project", in.ProjectID), map[string]any{"resource": "project", "action": "delete", "project": in.ProjectID}
		})
	}
	return "", nil, fmt.Errorf("suite: no scope rule for %q", name)
}
