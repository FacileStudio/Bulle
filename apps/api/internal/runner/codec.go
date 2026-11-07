package runner

import (
	"cmp"
	"encoding/json"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/nacelle"
)

// WirePart is one stored part as the session views read it: the fields
// nacelle's codec writes, plus Bulle's attachment reference.
type WirePart struct {
	Type     string          `json:"type"`
	Text     string          `json:"text,omitempty"`
	ID       string          `json:"id,omitempty"`
	Name     string          `json:"name,omitempty"`
	Input    json.RawMessage `json:"input,omitempty"`
	Finished bool            `json:"finished,omitempty"`
	Result   string          `json:"result,omitempty"`
	Failed   bool            `json:"failed,omitempty"`
	Stop     string          `json:"stop,omitempty"`
	Media    string          `json:"media_type,omitempty"`
	Size     int64           `json:"size,omitempty"`
}

// WireMessage is one stored message as the session views read it.
type WireMessage struct {
	Role  string     `json:"role"`
	Parts []WirePart `json:"parts"`
}

type wireMessage = WireMessage

// DecodeWire reads a stored message for display, with its tools under their
// current names and a call_tool call shown as the app tool it ran.
func DecodeWire(raw json.RawMessage) (WireMessage, error) {
	var wire WireMessage
	err := json.Unmarshal(raw, &wire)
	for i, part := range wire.Parts {
		if part.Type == "tool_call" {
			wire.Parts[i].Name, wire.Parts[i].Input = routed(part.Name, part.Input)
		}
	}
	canonicalTools(wire.Parts)
	return wire, err
}

// canonicalTools renames the tool calls and results a stored message keeps
// under a tool name an older Bulle or nacelle used, and their old argument
// names, so an old transcript replays against the tools the run declares now.
func canonicalTools(parts []WirePart) {
	legacy := nacelle.LegacyToolNames()
	for i, part := range parts {
		if part.Type != "tool_call" && part.Type != "tool_result" {
			continue
		}
		parts[i].Name = suite.CanonicalTool(cmp.Or(legacy[part.Name], part.Name))
		if part.Type == "tool_call" {
			parts[i].Input = suite.CanonicalInput(parts[i].Name, part.Input)
		}
	}
}

// MessageTexts reads a stored message as the separate things it says. The
// engine merges two adjacent messages of the same role into one message with a
// text part each, so a transcript written by an older build can hold a whole
// run of user turns in one row: the part boundaries are the only record left of
// where one message ended and the next began.
func MessageTexts(raw json.RawMessage) (role string, texts []string, err error) {
	var wire wireMessage
	if err := json.Unmarshal(raw, &wire); err != nil {
		return "", nil, err
	}
	texts = make([]string, 0, len(wire.Parts))
	for _, part := range wire.Parts {
		if part.Type == "text" && part.Text != "" && !IsStamp(part.Text) {
			texts = append(texts, part.Text)
		}
	}
	return wire.Role, texts, nil
}

// MessageText is those texts as one body, one paragraph each. Joining them with
// nothing glued "hello" and "how are you" into "hellohow are you".
func MessageText(raw json.RawMessage) (role, text string, err error) {
	role, texts, err := MessageTexts(raw)
	if err != nil {
		return "", "", err
	}
	return role, strings.Join(texts, "\n\n"), nil
}

// Redactor is the secrets a stored transcript and a streamed event must
// never carry.
type Redactor []string

// NewRedactor keeps the secrets long enough to redact without mangling
// ordinary words.
func NewRedactor(secrets ...string) Redactor {
	kept := make(Redactor, 0, len(secrets))
	for _, secret := range secrets {
		if len(secret) >= 8 {
			kept = append(kept, secret)
		}
	}
	return kept
}

// Apply replaces every secret in text.
func (r Redactor) Apply(text string) string {
	for _, secret := range r {
		text = strings.ReplaceAll(text, secret, "[redacted]")
	}
	return text
}
