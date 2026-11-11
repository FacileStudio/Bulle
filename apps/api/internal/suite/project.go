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
