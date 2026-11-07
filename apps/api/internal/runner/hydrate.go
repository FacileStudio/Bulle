package runner

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"strings"
	"sync"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

const maxFileText = 200 << 10

// hydrator swaps each attachment reference in the user's messages for the file
// itself just before a request leaves for the model, so the engine's
// conversation, and the transcript saved from it, only ever hold references.
// One hydrator lives for one run, and it loads each attachment once: every
// request and token count of the run reuses the part it built.
type hydrator struct {
	nacelle.Backend
	files  store.Attachments
	userID int64
	pdf    bool
	loaded *sync.Map
}

func (r *Runner) hydrating(backend nacelle.Backend, userID int64, backendName string) nacelle.Backend {
	return hydrator{Backend: backend, files: r.cfg.Store, userID: userID, pdf: backendName != "google", loaded: &sync.Map{}}
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
	return out, nil
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
	if cached, ok := h.loaded.Load(ref.ID); ok {
		return cached.(nacelle.Part), nil
	}
	loaded, err := h.load(ctx, ref)
	if err != nil {
		return nil, err
	}
	h.loaded.Store(ref.ID, loaded)
	return loaded, nil
}

func (h hydrator) load(ctx context.Context, ref AttachmentRef) (nacelle.Part, error) {
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
