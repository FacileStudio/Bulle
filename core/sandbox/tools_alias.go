package sandbox

import (
	"encoding/json"
	"math"
)

var pathAliases = map[string]string{"file_path": "path"}

var editAliases = map[string]string{"file_path": "path", "old": "old_string", "new": "new_string"}

// decodeAliased decodes data into target after renaming each alias key to the
// name the tool declares, the way nacelle's own tools read them. A canonical
// key the model also sent wins over its alias.
func decodeAliased(data []byte, target any, aliases map[string]string) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for from, to := range aliases {
		value, ok := fields[from]
		if !ok {
			continue
		}
		delete(fields, from)
		if _, taken := fields[to]; !taken {
			fields[to] = value
		}
	}
	canonical, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	return json.Unmarshal(canonical, target)
}

// UnmarshalJSON accepts file_path as an alias for path.
func (in *readInput) UnmarshalJSON(data []byte) error {
	type plain readInput
	return decodeAliased(data, (*plain)(in), pathAliases)
}

// UnmarshalJSON accepts file_path as an alias for path.
func (in *writeInput) UnmarshalJSON(data []byte) error {
	type plain writeInput
	return decodeAliased(data, (*plain)(in), pathAliases)
}

// UnmarshalJSON accepts file_path, old and new, the names edit_file took
// before nacelle v0.32.
func (in *editInput) UnmarshalJSON(data []byte) error {
	type plain editInput
	return decodeAliased(data, (*plain)(in), editAliases)
}

// UnmarshalJSON accepts timeout, in seconds, as the pre-v0.32 spelling of
// timeout_ms. timeout_ms wins when both are sent.
func (in *commandInput) UnmarshalJSON(data []byte) error {
	type plain commandInput
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
