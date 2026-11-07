package runner

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/FacileStudio/nacelle"
)

func TestStoredMessagesKeepImagesReasoningAndNative(t *testing.T) {
	assistant := nacelle.Message{Role: nacelle.RoleAssistant, Parts: []nacelle.Part{
		nacelle.Reasoning{Text: "thinking it over"},
		nacelle.Native{Provider: "anthropic", Data: json.RawMessage(`{"type":"thinking","signature":"sig"}`)},
		nacelle.ToolCall{ID: "c1", Name: "browser_screenshot", Input: json.RawMessage(`{}`), Finished: true},
		nacelle.Finish{Stop: nacelle.StopTools},
	}}
	results := nacelle.Message{Role: nacelle.RoleUser, Parts: []nacelle.Part{
		nacelle.ToolResult{ID: "c1", Name: "browser_screenshot", Result: "a page", Images: []nacelle.Image{{MediaType: "image/png", Data: []byte("png")}}},
	}}
	for _, message := range []nacelle.Message{assistant, results} {
		raw, err := EncodeMessage(message, nil)
		if err != nil {
			t.Fatal(err)
		}
		back, err := DecodeMessage(raw)
		if err != nil || !reflect.DeepEqual(back, message) {
			t.Fatalf("round trip lost a part:\n got %+v\nwant %+v (%v)", back, message, err)
		}
	}
}

func TestStoredMessagesKeepAttachmentsAndRedact(t *testing.T) {
	ref := AttachmentRef{ID: "a1", Name: "doc.pdf", MediaType: "application/pdf", Size: 3}
	secret := "sk-very-secret-token"
	message := nacelle.Message{Role: nacelle.RoleUser, Parts: []nacelle.Part{AttachmentPart(ref), nacelle.Text{Text: "key " + secret}}}
	raw, err := EncodeMessage(message, NewRedactor(secret))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), secret) || !strings.Contains(string(raw), `"type":"attachment"`) {
		t.Fatalf("stored %s", raw)
	}
	back, err := DecodeMessage(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := attachmentOf(back.Parts[0].(nacelle.Text).Text); !ok || got != ref {
		t.Fatalf("the attachment came back as %+v", back.Parts[0])
	}
	signed := nacelle.Message{Role: nacelle.RoleAssistant, Parts: []nacelle.Part{
		nacelle.Native{Provider: "anthropic", Data: json.RawMessage(`{"thinking":"` + secret + `"}`)},
		nacelle.Text{Text: "done"},
	}}
	raw, _ = EncodeMessage(signed, NewRedactor(secret))
	if strings.Contains(string(raw), secret) || strings.Contains(string(raw), "native") {
		t.Fatalf("a Native part holding a secret must be dropped: %s", raw)
	}
}
