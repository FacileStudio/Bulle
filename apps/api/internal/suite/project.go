package suite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func unwrapData(raw []byte) []byte {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return trimmed
	}
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(trimmed, &envelope); err != nil || len(envelope.Data) == 0 {
		return trimmed
	}
	return envelope.Data
}

func decodeInto(raw []byte, target any) error {
	return json.Unmarshal(unwrapData(raw), target)
}

func pickMap(raw []byte, keys ...string) (map[string]any, error) {
	var item map[string]any
	if err := decodeInto(raw, &item); err != nil {
		return nil, err
	}
	kept := make(map[string]any, len(keys))
	for _, key := range keys {
		if value, ok := item[key]; ok {
			kept[key] = value
		}
	}
	return kept, nil
}

func pickOne(raw []byte, keys ...string) ([]byte, error) {
	kept, err := pickMap(raw, keys...)
	if err != nil {
		return nil, err
	}
	return json.Marshal(kept)
}

func toMap(value any) (map[string]any, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	out := map[string]any{}
	return out, json.Unmarshal(data, &out)
}

func parseWhen(value string, today time.Time, location *time.Location) (time.Time, error) {
	value = strings.TrimSpace(value)
	switch strings.ToLower(value) {
	case "", "today":
		return today, nil
	case "tomorrow":
		return today.AddDate(0, 0, 1), nil
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}
	if parsed, err := time.ParseInLocation("2006-01-02", value, location); err == nil {
		return parsed, nil
	}
	return time.Time{}, fmt.Errorf("%q is not a date: use today, tomorrow, YYYY-MM-DD or RFC 3339", value)
}
