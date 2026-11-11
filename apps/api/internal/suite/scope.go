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
	case strings.HasPrefix(name, "glouton_"):
		return scopeGlouton(name, input, d)
	case strings.HasPrefix(name, "sablier_"):
		return scopeSablier(name, input, d)
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
