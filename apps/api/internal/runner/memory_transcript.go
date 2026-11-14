package runner

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

const memoryTranscriptLimit = 24 << 10

// passTranscript is the user and assistant text of messages from seq from on,
// one labelled paragraph per message, cut at a message boundary near 24 KB.
// It answers the seq right after the last message it read, so the next pass
// starts there. From the first message that brought a third party's text into
// the chat on, earlier ones included, the assistant's text is left out: it
// may repeat what a mail or a page said, so only the owner's own words reach
// memory.
func passTranscript(messages []store.Message, from int) (string, int) {
	var b strings.Builder
	upto := from
	tainted := false
	for _, message := range messages {
		tainted = tainted || taintsChat(message.Payload)
		if message.Seq < from {
			continue
		}
		entry := transcriptEntry(message, tainted)
		if b.Len() > 0 && b.Len()+len(entry) > memoryTranscriptLimit {
			break
		}
		b.WriteString(clipEntry(entry))
		upto = message.Seq + 1
	}
	return strings.TrimSpace(b.String()), upto
}

func transcriptEntry(message store.Message, ownOnly bool) string {
	role, texts := store.ChatTexts(message.Payload)
	if len(texts) == 0 || (ownOnly && role != "user") {
		return ""
	}
	label := "User"
	if role == "assistant" {
		label = "Assistant"
	}
	return label + ": " + strings.Join(texts, "\n\n") + "\n\n"
}

func clipEntry(entry string) string {
	if len(entry) <= memoryTranscriptLimit {
		return entry
	}
	return strings.ToValidUTF8(entry[:memoryTranscriptLimit], "") + "\n[message truncated]\n\n"
}

// taintsChat reports whether a stored message brought untrusted text into
// the chat: a call or a result of an untrusted source, or an attachment. A
// message that does not decode counts as one.
func taintsChat(payload json.RawMessage) bool {
	message, err := DecodeMessage(payload)
	if err != nil {
		return true
	}
	return slices.ContainsFunc(message.Parts, func(part nacelle.Part) bool {
		switch part := part.(type) {
		case nacelle.ToolCall:
			return untrusted(part.Name)
		case nacelle.ToolResult:
			return untrusted(part.Name)
		case nacelle.Text:
			_, attached := attachmentOf(part.Text)
			return attached
		}
		return false
	})
}
