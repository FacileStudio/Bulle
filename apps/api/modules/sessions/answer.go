package sessions

import (
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
)

// lastText is the turn's final text part: the answer a model gives after its
// last tool call, without the narration it wrote between steps.
func lastText(texts []string) string {
	if len(texts) == 0 {
		return ""
	}
	return texts[len(texts)-1]
}

// userMessages is one stored user row as the bubbles it shows. A row merged
// from several turns keeps one bubble per text part; its attachments go on the
// last bubble, and a message of attachments alone is still a bubble.
func userMessages(parts []runner.WirePart) []StoredMessage {
	var texts []string
	var files []AttachmentInfo
	for _, part := range parts {
		switch {
		case part.Type == "text" && part.Text != "":
			texts = append(texts, part.Text)
		case part.Type == "attachment":
			ref := part.Attachment()
			files = append(files, AttachmentInfo{ID: ref.ID, Name: ref.Name, MediaType: ref.MediaType, Size: ref.Size})
		}
	}
	if len(texts) == 0 && len(files) == 0 {
		return nil
	}
	if len(texts) < 2 {
		texts = []string{strings.Join(texts, "\n\n")}
	}
	out := make([]StoredMessage, len(texts))
	for i, text := range texts {
		out[i] = StoredMessage{Role: "user", Content: text}
	}
	out[len(out)-1].Attachments = files
	return out
}
