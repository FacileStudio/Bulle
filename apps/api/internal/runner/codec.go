package runner

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/FacileStudio/nacelle"
)

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
}

type WireMessage struct {
	Role  string     `json:"role"`
	Parts []WirePart `json:"parts"`
}

type wirePart = WirePart
type wireMessage = WireMessage

func DecodeWire(raw json.RawMessage) (WireMessage, error) {
	var wire WireMessage
	err := json.Unmarshal(raw, &wire)
	return wire, err
}

func EncodeMessage(message nacelle.Message, redact Redactor) (json.RawMessage, error) {
	wire := wireMessage{Role: string(message.Role), Parts: make([]wirePart, 0, len(message.Parts))}
	for _, part := range message.Parts {
		switch p := part.(type) {
		case nacelle.Text:
			wire.Parts = append(wire.Parts, wirePart{Type: "text", Text: redact.Apply(p.Text)})
		case nacelle.Reasoning:
			wire.Parts = append(wire.Parts, wirePart{Type: "reasoning", Text: redact.Apply(p.Text)})
		case nacelle.ToolCall:
			input := p.Input
			if !json.Valid(input) {
				input = nil
			}
			wire.Parts = append(wire.Parts, wirePart{Type: "tool_call", ID: p.ID, Name: p.Name, Input: input, Finished: p.Finished})
		case nacelle.ToolResult:
			wire.Parts = append(wire.Parts, wirePart{Type: "tool_result", ID: p.ID, Name: p.Name, Result: redact.Apply(p.Result), Failed: p.Failed})
		case nacelle.Finish:
			wire.Parts = append(wire.Parts, wirePart{Type: "finish", Stop: string(p.Stop)})
		default:
			return nil, fmt.Errorf("runner: unknown message part %T", part)
		}
	}
	return json.Marshal(wire)
}

func DecodeMessage(raw json.RawMessage) (nacelle.Message, error) {
	var wire wireMessage
	if err := json.Unmarshal(raw, &wire); err != nil {
		return nacelle.Message{}, err
	}
	message := nacelle.Message{Role: nacelle.Role(wire.Role), Parts: make([]nacelle.Part, 0, len(wire.Parts))}
	for _, part := range wire.Parts {
		switch part.Type {
		case "text":
			message.Parts = append(message.Parts, nacelle.Text{Text: part.Text})
		case "reasoning":
			message.Parts = append(message.Parts, nacelle.Reasoning{Text: part.Text})
		case "tool_call":
			message.Parts = append(message.Parts, nacelle.ToolCall{ID: part.ID, Name: part.Name, Input: part.Input, Finished: part.Finished})
		case "tool_result":
			message.Parts = append(message.Parts, nacelle.ToolResult{ID: part.ID, Name: part.Name, Result: part.Result, Failed: part.Failed})
		case "finish":
			message.Parts = append(message.Parts, nacelle.Finish{Stop: nacelle.Stop(part.Stop)})
		default:
			return nacelle.Message{}, fmt.Errorf("runner: unknown stored part type %q", part.Type)
		}
	}
	return message, nil
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
		if part.Type == "text" && part.Text != "" {
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

type Redactor []string

func NewRedactor(secrets ...string) Redactor {
	kept := make(Redactor, 0, len(secrets))
	for _, secret := range secrets {
		if len(secret) >= 8 {
			kept = append(kept, secret)
		}
	}
	return kept
}

func (r Redactor) Apply(text string) string {
	for _, secret := range r {
		text = strings.ReplaceAll(text, secret, "[redacted]")
	}
	return text
}
