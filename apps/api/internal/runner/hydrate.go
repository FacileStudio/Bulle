package runner

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"strings"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

const maxFileText = 200 << 10

// hydrator swaps each attachment reference in the user's messages for the file
// itself just before a request leaves for the model, so the engine's
// conversation, and the transcript saved from it, only ever hold references.
// It also stamps the run's prompt with the date and time the run started, read
// once so every request of one run carries the same stamp and the prompt cache
// holds across the run's tool calls.
type hydrator struct {
	nacelle.Backend
	files  store.Attachments
	userID int64
	pdf    bool
	now    time.Time
}

func (r *Runner) hydrating(backend nacelle.Backend, userID int64, backendName string) nacelle.Backend {
	return hydrator{Backend: backend, files: r.cfg.Store, userID: userID, pdf: backendName != "google", now: time.Now()}
}

func (h hydrator) Stream(ctx context.Context, request nacelle.Request) iter.Seq2[nacelle.Event, error] {
	messages, err := h.hydrate(ctx, request.Messages)
	if err != nil {
		return func(yield func(nacelle.Event, error) bool) { yield(nacelle.Event{}, err) }
	}
	request.Messages = messages
	return h.Backend.Stream(ctx, request)
}

func (h hydrator) CountTokens(ctx context.Context, request nacelle.Request) (int64, error) {
	messages, err := h.hydrate(ctx, request.Messages)
	if err != nil {
		return 0, err
	}
	request.Messages = messages
	return h.Backend.CountTokens(ctx, request)
}

func (h hydrator) hydrate(ctx context.Context, messages []nacelle.Message) ([]nacelle.Message, error) {
	out := make([]nacelle.Message, len(messages))
	for i, message := range messages {
		out[i] = message
		if message.Role != nacelle.RoleUser {
			continue
		}
		parts := make([]nacelle.Part, len(message.Parts))
		for j, part := range message.Parts {
			hydrated, err := h.part(ctx, part)
			if err != nil {
				return nil, err
			}
			parts[j] = hydrated
		}
		out[i].Parts = parts
	}
	stamp(out, h.now)
	return out, nil
}

// stamp puts the clock in front of the newest user message that has text, which
// is the run's prompt: tool results are user messages too, but carry no text.
func stamp(messages []nacelle.Message, now time.Time) {
	if now.IsZero() {
		return
	}
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role != nacelle.RoleUser || !hasText(messages[i]) {
			continue
		}
		clock := nacelle.Text{Text: "[Current date and time: " + now.Format("Monday 2 January 2006, 15:04 MST (-07:00)") + "]"}
		messages[i].Parts = append([]nacelle.Part{clock}, messages[i].Parts...)
		return
	}
}

func hasText(message nacelle.Message) bool {
	for _, part := range message.Parts {
		if _, ok := part.(nacelle.Text); ok {
			return true
		}
	}
	return false
}

func (h hydrator) part(ctx context.Context, part nacelle.Part) (nacelle.Part, error) {
	text, ok := part.(nacelle.Text)
	if !ok {
		return part, nil
	}
	ref, ok := attachmentOf(text.Text)
	if !ok {
		return part, nil
	}
	found, err := h.files.GetAttachment(ctx, h.userID, ref.ID)
	if errors.Is(err, store.ErrNotFound) {
		return nacelle.Text{Text: fmt.Sprintf("[The attachment %q is no longer available.]", ref.Name)}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("loading attachment %s: %w", ref.Name, err)
	}
	return h.file(found), nil
}

func (h hydrator) file(found store.Attachment) nacelle.Part {
	switch {
	case strings.HasPrefix(found.MediaType, "image/"):
		return nacelle.Image{MediaType: found.MediaType, Data: found.Data}
	case found.MediaType == "application/pdf" && h.pdf:
		return nacelle.Document{Name: found.Name, MediaType: found.MediaType, Data: found.Data}
	case found.MediaType == "application/pdf":
		return nacelle.Text{Text: fmt.Sprintf("[The user attached the PDF %q, but this model cannot read PDFs. Tell them to switch to another model to read it.]", found.Name)}
	}
	return FileText(found.Name, found.Data)
}

// FileText is a text attachment as the model reads it, capped at 200 KB.
func FileText(name string, data []byte) nacelle.Text {
	note := ""
	if len(data) > maxFileText {
		data = data[:maxFileText]
		note = "\n[truncated at 200 KB]"
	}
	body := strings.ToValidUTF8(string(data), "")
	return nacelle.Text{Text: fmt.Sprintf("<file name=%q>\n%s%s\n</file>", name, body, note)}
}
