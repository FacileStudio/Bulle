package suite

import (
	"bytes"
	"encoding/json"
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

func pick(raw []byte, keys ...string) ([]byte, error) {
	var items []map[string]json.RawMessage
	if err := decodeInto(raw, &items); err != nil {
		return nil, err
	}
	out := make([]map[string]json.RawMessage, 0, len(items))
	for _, item := range items {
		kept := make(map[string]json.RawMessage, len(keys))
		for _, key := range keys {
			if value, ok := item[key]; ok {
				kept[key] = value
			}
		}
		out = append(out, kept)
	}
	return json.Marshal(out)
}

func appendJSONList(base, next []byte) ([]byte, error) {
	var left, right []json.RawMessage
	if len(base) > 0 {
		if err := json.Unmarshal(base, &left); err != nil {
			return nil, err
		}
	}
	if err := decodeInto(next, &right); err != nil {
		return nil, err
	}
	return json.Marshal(append(left, right...))
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
