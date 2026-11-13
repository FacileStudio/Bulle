package runner

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/FacileStudio/nacelle"
)

const (
	MaxAttachments   = 10
	attachmentMarker = "\x00bulle:attachment\x00"
)

var (
	ErrTooManyAttachments = errors.New("runner: too many attachments")
	ErrUnknownAttachment  = errors.New("runner: unknown attachment")
)

// AttachmentRef is how a message points at an uploaded file. The transcript
// keeps the reference and never the bytes, which the runner loads again each
// time it sends the conversation to a model.
type AttachmentRef struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	MediaType string `json:"media_type"`
	Size      int64  `json:"size"`
}

// AttachmentPart carries a reference through the engine's conversation as a
// text part the hydrating backend recognises. User text cannot forge one,
// because StartWithAttachments strips the NUL bytes the marker is made of.
func AttachmentPart(ref AttachmentRef) nacelle.Text {
	raw, _ := json.Marshal(ref)
	return nacelle.Text{Text: attachmentMarker + string(raw)}
}

func attachmentOf(text string) (AttachmentRef, bool) {
	rest, ok := strings.CutPrefix(text, attachmentMarker)
	if !ok {
		return AttachmentRef{}, false
	}
	var ref AttachmentRef
	if json.Unmarshal([]byte(rest), &ref) != nil || ref.ID == "" {
		return AttachmentRef{}, false
	}
	return ref, true
}

// Attachment reads a stored attachment part back as its reference.
func (p WirePart) Attachment() AttachmentRef {
	return AttachmentRef{ID: p.ID, Name: p.Name, MediaType: p.Media, Size: p.Size}
}

// StartWithAttachments is Start with files attached. The text may be empty
// when at least one attachment is given. Every id must name one of the user's
// own uploads.
func (r *Runner) StartWithAttachments(ctx context.Context, userID int64, sessionID, text string, ids []string) (string, error) {
	return r.StartPrompt(ctx, userID, sessionID, Prompt{Text: text, Attachments: ids})
}

// Prompt is one message a person sends: its text, the ids of the files
// attached to it, and the suite page open beside the chat, if any.
type Prompt struct {
	Text        string
	Attachments []string
	Page        PageContext
}

// StartPrompt is StartWithAttachments that may come from a chat a suite app
// embeds beside one of its pages: the model reads which page is open in
// front of the message, and the person never sees that note. A zero page
// adds nothing.
func (r *Runner) StartPrompt(ctx context.Context, userID int64, sessionID string, prompt Prompt) (string, error) {
	text, page := strings.TrimSpace(strings.ReplaceAll(prompt.Text, "\x00", "")), prompt.Page
	refs, err := r.attachmentRefs(ctx, userID, prompt.Attachments)
	if err != nil {
		return "", err
	}
	if text == "" && len(refs) == 0 {
		return "", ErrEmptyMessage
	}
	message := nacelle.Message{Role: nacelle.RoleUser}
	if !page.IsZero() {
		note, err := page.Note()
		if err != nil {
			return "", err
		}
		message.Parts = append(message.Parts, note)
	}
	for _, ref := range refs {
		message.Parts = append(message.Parts, AttachmentPart(ref))
	}
	title := text
	if text != "" {
		message.Parts = append(message.Parts, nacelle.Text{Text: text})
	} else {
		title = refs[0].Name
	}
	return r.start(ctx, userID, sessionID, message, title)
}

func (r *Runner) attachmentRefs(ctx context.Context, userID int64, ids []string) ([]AttachmentRef, error) {
	if len(ids) > MaxAttachments {
		return nil, ErrTooManyAttachments
	}
	if len(ids) == 0 {
		return nil, nil
	}
	found, err := r.cfg.Store.AttachmentInfo(ctx, userID, ids)
	if err != nil {
		return nil, err
	}
	if len(found) != len(ids) {
		return nil, ErrUnknownAttachment
	}
	refs := make([]AttachmentRef, len(found))
	for i, a := range found {
		refs[i] = AttachmentRef{ID: a.ID, Name: a.Name, MediaType: a.MediaType, Size: a.Size}
	}
	return refs, nil
}

func (r *Runner) start(ctx context.Context, userID int64, sessionID string, message nacelle.Message, title string) (string, error) {
	if err := r.sessionAccess(ctx, userID, sessionID); err != nil {
		return "", err
	}
	slot, err := r.slotFor(ctx, userID, sessionID)
	if err != nil {
		return "", err
	}
	if err := r.reserve(slot); err != nil {
		return "", err
	}
	runID, session, history, err := r.prepare(ctx, userID, sessionID, message, title)
	if err != nil {
		r.release(slot)
		return "", err
	}

	runCtx, cancel := context.WithCancelCause(r.base)
	em := &emitter{store: r.cfg.Store, hub: r.hub, logger: r.cfg.Logger, runID: runID}
	lr := &liveRun{id: runID, userID: userID, session: session, cancel: cancel, emitter: em, started: r.now()}
	lr.gate = newGate(r.outward, r.cfg.Store, em, runID, r.now)
	lr.gate.pended = r.approvalPending(lr)
	lr.gate.decided = r.approvalDecided(lr)
	lr.gate.prime(r.cfg.Store, userID, history)

	r.mu.Lock()
	r.live[runID] = lr
	r.mu.Unlock()
	r.announceStart(ctx, lr)
	r.recordPresence(ctx, userID, session)
	go r.execute(runCtx, lr, session, history)
	return runID, nil
}
