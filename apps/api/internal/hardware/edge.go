package hardware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// EdgeTool is one tool a device advertised in its tools/list answer.
type EdgeTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// Discover asks the device for its tools and keeps the answer.
func (p *Peer) Discover(ctx context.Context) error {
	result, err := p.Call(ctx, "tools/list", nil)
	if err != nil {
		return err
	}
	var list struct {
		Tools []EdgeTool `json:"tools"`
	}
	if err := json.Unmarshal(result, &list); err != nil {
		return fmt.Errorf("hardware: tools/list answer: %w", err)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.tools, p.ready = list.Tools, true
	return nil
}

// Tools answers the tools the device advertised, and false until it has.
func (p *Peer) Tools() ([]EdgeTool, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.tools, p.ready && !p.closed
}

func resultText(result json.RawMessage) (string, error) {
	var out struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	}
	if err := json.Unmarshal(result, &out); err != nil || len(out.Content) == 0 {
		return string(result), nil
	}
	var parts []string
	for _, c := range out.Content {
		if c.Type == "text" {
			parts = append(parts, c.Text)
		}
	}
	text := strings.Join(parts, "\n")
	if out.IsError {
		return "", errors.New(text)
	}
	return text, nil
}

func toolName(device Device, remote string, taken map[string]bool) string {
	name := slug(device.Name) + "__" + slug(remote)
	if len(name) > maxToolName {
		name = name[:maxToolName]
	}
	if !taken[name] {
		return name
	}
	suffix := "_" + strings.TrimPrefix(device.ID, IDPrefix)[:6]
	return name[:min(len(name), maxToolName-len(suffix))] + suffix
}

func slug(s string) string {
	var b strings.Builder
	underscore := false
	for _, r := range strings.ToLower(s) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
			underscore = false
			continue
		}
		if !underscore && b.Len() > 0 {
			b.WriteByte('_')
			underscore = true
		}
	}
	out := strings.TrimSuffix(b.String(), "_")
	if out == "" {
		return "device"
	}
	return out
}

var schemaTypes = map[any]bool{
	"string": true, "integer": true, "number": true, "boolean": true, "array": true, "object": true, "null": true,
}

// objectSchema answers the input schema a model API accepts, and false for one
// that would make the provider refuse the whole request: a device must not be
// able to break every run of its owner.
func objectSchema(schema map[string]any) (map[string]any, bool) {
	if schema == nil {
		return map[string]any{"type": "object", "properties": map[string]any{}}, true
	}
	if schema["type"] != "object" {
		return nil, false
	}
	props, ok := schema["properties"].(map[string]any)
	if !ok && schema["properties"] != nil {
		return nil, false
	}
	for _, prop := range props {
		field, ok := prop.(map[string]any)
		if !ok || (field["type"] != nil && !schemaTypes[field["type"]]) {
			return nil, false
		}
	}
	return schema, true
}

func oneLine(text string, limit int) string {
	text = strings.Join(strings.Fields(text), " ")
	if runes := []rune(text); len(runes) > limit {
		return string(runes[:limit]) + "..."
	}
	return text
}
