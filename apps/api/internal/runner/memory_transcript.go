package runner

import (
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

const memoryTranscriptLimit = 24 << 10

// passTranscript is the user and assistant text of messages from seq from on,
// one labelled paragraph per message, cut at a message boundary near 24 KB.
// It answers the seq right after the last message it read, so the next pass
// starts there.
func passTranscript(messages []store.Message, from int) (string, int) {
	var b strings.Builder
	upto := from
	for _, message := range messages {
		if message.Seq < from {
			continue
		}
		entry := transcriptEntry(message)
		if b.Len() > 0 && b.Len()+len(entry) > memoryTranscriptLimit {
			break
		}
		b.WriteString(clipEntry(entry))
		upto = message.Seq + 1
	}
	return strings.TrimSpace(b.String()), upto
}

func transcriptEntry(message store.Message) string {
	role, texts := store.ChatTexts(message.Payload)
	if len(texts) == 0 {
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
