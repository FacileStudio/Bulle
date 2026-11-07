package runner

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/FacileStudio/nacelle"
)

// Access decides whether a user may start AI work. A nil Access allows everyone.
// Its refusal is returned to the caller unchanged, so it can carry its own
// HTTP status.
type Access interface {
	CheckAI(ctx context.Context, userID int64) error
}

func (r *Runner) checkAccess(ctx context.Context, userID int64) error {
	if r.cfg.Access == nil {
		return nil
	}
	return r.cfg.Access.CheckAI(ctx, userID)
}

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
	text = strings.TrimSpace(strings.ReplaceAll(text, "\x00", ""))
	refs, err := r.attachmentRefs(ctx, userID, ids)
	if err != nil {
		return "", err
	}
	if text == "" && len(refs) == 0 {
		return "", ErrEmptyMessage
	}
	message := nacelle.Message{Role: nacelle.RoleUser}
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
	if err := r.checkAccess(ctx, userID); err != nil {
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
