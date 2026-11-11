package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/google/jsonschema-go/jsonschema"
)

func resolveSchema(raw json.RawMessage) (*jsonschema.Resolved, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	if len(raw) > maxCompleteSchema {
		return nil, troncerrors.TooLarge("schema is larger than 8 KiB")
	}
	var schema jsonschema.Schema
	if err := json.Unmarshal(raw, &schema); err != nil {
		return nil, troncerrors.Invalid("schema must be a JSON Schema object")
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		return nil, troncerrors.Invalid("schema is not a valid JSON Schema without remote references")
	}
	return resolved, nil
}

// parseCompletion returns the text as is, or, with a schema, the one JSON
// value the model wrote (possibly in a code fence) once it validates.
func parseCompletion(raw, model string, schema *jsonschema.Resolved, err error) (CompleteResponse, error) {
	if err != nil {
		return CompleteResponse{Model: model}, err
	}
	if schema == nil {
		if strings.TrimSpace(raw) == "" {
			return CompleteResponse{Model: model}, errors.New("completion is empty")
		}
		return CompleteResponse{Text: raw, Model: model}, nil
	}
	value, err := decodeJSONValue(raw)
	if err != nil {
		return CompleteResponse{Model: model}, err
	}
	var instance any
	if err := json.Unmarshal(value, &instance); err != nil {
		return CompleteResponse{Model: model}, err
	}
	if err := schema.Validate(instance); err != nil {
		return CompleteResponse{Model: model}, fmt.Errorf("completion does not match the schema: %w", err)
	}
	return CompleteResponse{JSON: value, Model: model}, nil
}

func decodeJSONValue(raw string) (json.RawMessage, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(raw, "```json"), "```"), "```")
	decoder := json.NewDecoder(strings.NewReader(strings.TrimSpace(raw)))
	var value json.RawMessage
	if err := decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("completion is not JSON: %w", err)
	}
	if decoder.More() {
		return nil, errors.New("completion has trailing data")
	}
	return compactRaw(value), nil
}

func compactRaw(value json.RawMessage) json.RawMessage {
	var out bytes.Buffer
	if err := json.Compact(&out, value); err != nil {
		return value
	}
	return out.Bytes()
}
