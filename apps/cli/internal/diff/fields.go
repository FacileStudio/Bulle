package diff

import "encoding/json"

// fieldString reads one string argument out of already-parsed input.
func fieldString(fields map[string]json.RawMessage, key string) (string, bool) {
	raw, ok := fields[key]
	if !ok {
		return "", false
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", false
	}
	return value, true
}

// fieldOf reads a string argument under its name, or under the alias an
// older edit_file used for it.
func fieldOf(fields map[string]json.RawMessage, name, alias string) (string, bool) {
	if value, ok := fieldString(fields, name); ok {
		return value, true
	}
	return fieldString(fields, alias)
}
