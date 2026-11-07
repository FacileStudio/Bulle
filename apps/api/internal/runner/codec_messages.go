package runner

import (
	"encoding/json"

	"github.com/FacileStudio/nacelle"
)

// storedMessage is a message as a row holds it, with each part left raw so
// Bulle's own attachment part and the tool renames can be applied around
// nacelle's codec.
type storedMessage struct {
	Role  string            `json:"role"`
	Parts []json.RawMessage `json:"parts"`
}

// EncodeMessage stores a message with nacelle's codec, so reasoning, tool
// images and provider Native parts survive, after redacting the secrets in
// its text. A user attachment is stored as its reference, never its bytes.
func EncodeMessage(message nacelle.Message, redact Redactor) (json.RawMessage, error) {
	kept := nacelle.Message{Role: message.Role, Parts: make([]nacelle.Part, 0, len(message.Parts))}
	for _, part := range message.Parts {
		if clean, ok := redactPart(part, redact); ok {
			kept.Parts = append(kept.Parts, clean)
		}
	}
	raw, err := json.Marshal(kept)
	if err != nil || message.Role != nacelle.RoleUser {
		return raw, err
	}
	var stored storedMessage
	if err := json.Unmarshal(raw, &stored); err != nil {
		return nil, err
	}
	for i, part := range kept.Parts {
		text, isText := part.(nacelle.Text)
		if ref, ok := attachmentOf(text.Text); isText && ok {
			stored.Parts[i], _ = json.Marshal(WirePart{Type: "attachment", ID: ref.ID, Name: ref.Name, Media: ref.MediaType, Size: ref.Size})
		}
	}
	return json.Marshal(stored)
}

// redactPart replaces the secrets in a part's text. A Native part that holds
// one is dropped instead: its bytes are a provider's signed block, which an
// edit would invalidate, and the provider replays the turn without it.
func redactPart(part nacelle.Part, redact Redactor) (nacelle.Part, bool) {
	switch p := part.(type) {
	case nacelle.Text:
		p.Text = redact.Apply(p.Text)
		return p, true
	case nacelle.Reasoning:
		p.Text = redact.Apply(p.Text)
		return p, true
	case nacelle.ToolResult:
		p.Result = redact.Apply(p.Result)
		return p, true
	case nacelle.Native:
		return p, redact.Apply(string(p.Data)) == string(p.Data)
	}
	return part, true
}

// DecodeMessage reads a stored message back with nacelle's codec. An
// attachment part becomes the text part the hydrating backend recognises, and
// a tool an older Bulle or nacelle named differently takes its current name
// and argument names.
func DecodeMessage(raw json.RawMessage) (nacelle.Message, error) {
	var stored storedMessage
	if err := json.Unmarshal(raw, &stored); err != nil {
		return nacelle.Message{}, err
	}
	for i, part := range stored.Parts {
		current, err := currentPart(part)
		if err != nil {
			return nacelle.Message{}, err
		}
		stored.Parts[i] = current
	}
	patched, err := json.Marshal(stored)
	if err != nil {
		return nacelle.Message{}, err
	}
	var message nacelle.Message
	err = json.Unmarshal(patched, &message)
	return message, err
}

// currentPart is one stored part as nacelle's codec reads it today.
func currentPart(raw json.RawMessage) (json.RawMessage, error) {
	var wire WirePart
	if err := json.Unmarshal(raw, &wire); err != nil {
		return nil, err
	}
	switch wire.Type {
	case "attachment":
		return json.Marshal(map[string]string{"type": "text", "text": AttachmentPart(wire.Attachment()).Text})
	case "tool_call", "tool_result":
		return renamedTool(raw, wire)
	}
	return raw, nil
}

// renamedTool rewrites a stored tool part's name and input to the current
// ones, keeping every other field, images included, as stored.
func renamedTool(raw json.RawMessage, wire WirePart) (json.RawMessage, error) {
	parts := []WirePart{wire}
	canonicalTools(parts)
	if parts[0].Name == wire.Name && string(parts[0].Input) == string(wire.Input) {
		return raw, nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	fields["name"], _ = json.Marshal(parts[0].Name)
	if len(parts[0].Input) > 0 {
		fields["input"] = parts[0].Input
	}
	return json.Marshal(fields)
}
