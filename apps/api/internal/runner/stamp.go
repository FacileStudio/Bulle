package runner

import (
	"strings"
	"time"

	"github.com/FacileStudio/nacelle"
)

const stampPrefix = "[Current date and time: "

// stampPrompt puts the clock in front of the run's prompt before it is stored.
// Writing it once into the stored message, rather than into whichever message
// is newest at request time, keeps every earlier request byte for byte the same,
// so the provider's prompt cache holds across runs.
func stampPrompt(message nacelle.Message, now time.Time) nacelle.Message {
	if now.IsZero() {
		return message
	}
	clock := nacelle.Text{Text: stampPrefix + now.Format("Monday 2 January 2006, 15:04 MST (-07:00)") + "]"}
	message.Parts = append([]nacelle.Part{clock}, message.Parts...)
	return message
}

// IsStamp reports whether a stored text part is the clock the runner wrote in
// front of a prompt, which the model reads and a person should not see.
func IsStamp(text string) bool {
	return strings.HasPrefix(text, stampPrefix) && strings.HasSuffix(text, "]") && !strings.Contains(text, "\n")
}
