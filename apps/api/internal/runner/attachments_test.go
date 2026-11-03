package runner_test

import (
	"bytes"
	"context"
	"errors"
	"iter"
	"strings"
	"sync"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

type captureBackend struct {
	mu   sync.Mutex
	seen []nacelle.Message
}

func (c *captureBackend) Name() string { return "capture" }

func (c *captureBackend) Capabilities() nacelle.Capabilities {
	return nacelle.Capabilities{Thinking: true}
}

func (c *captureBackend) CountTokens(context.Context, nacelle.Request) (int64, error) { return 0, nil }

func (c *captureBackend) Stream(_ context.Context, req nacelle.Request) iter.Seq2[nacelle.Event, error] {
	c.mu.Lock()
	c.seen = req.Messages
	c.mu.Unlock()
	return func(yield func(nacelle.Event, error) bool) {
		if yield(nacelle.Event{Kind: nacelle.KindText, Text: "seen"}, nil) {
			yield(nacelle.Event{Kind: nacelle.KindDone, Stop: nacelle.StopEnd}, nil)
		}
	}
}

func (c *captureBackend) lastUser() []nacelle.Part {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := len(c.seen) - 1; i >= 0; i-- {
		if c.seen[i].Role == nacelle.RoleUser {
			return c.seen[i].Parts
		}
	}
	return nil
}

func upload(t *testing.T, st *store.Memory, userID int64, name string, file nacelle.Image) string {
	t.Helper()
	created, err := st.CreateAttachment(context.Background(), store.Attachment{
		ID: store.NewID(), UserID: userID, Name: name, MediaType: file.MediaType, Size: int64(len(file.Data)), Data: file.Data,
	})
	if err != nil {
		t.Fatal(err)
	}
	return created.ID
}

func TestAttachmentPartsRoundTripWithoutBytes(t *testing.T) {
	ref := runner.AttachmentRef{ID: "a1", Name: "notes.md", MediaType: "text/markdown", Size: 12}
	message := nacelle.Message{Role: nacelle.RoleUser, Parts: []nacelle.Part{runner.AttachmentPart(ref), nacelle.Text{Text: "read this"}}}
	raw, err := runner.EncodeMessage(message, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"type":"attachment","id":"a1","name":"notes.md","media_type":"text/markdown","size":12}`
	if !strings.Contains(string(raw), want) || strings.Contains(string(raw), "\\u0000") {
		t.Fatalf("stored message: %s", raw)
	}
	back, err := runner.DecodeMessage(raw)
	if err != nil || len(back.Parts) != 2 || back.Parts[0] != message.Parts[0] || back.Parts[1] != message.Parts[1] {
		t.Fatalf("round trip: %v %+v", err, back)
	}
	_, texts, _ := runner.MessageTexts(raw)
	if len(texts) != 1 || texts[0] != "read this" {
		t.Fatalf("texts: %q", texts)
	}
}

func TestHistoryHydratesAttachments(t *testing.T) {
	backend := &captureBackend{}
	f := newFixture(t, backend, nil, nil)
	png := []byte("\x89PNG\r\n\x1a\nimage")
	ids := []string{
		upload(t, f.store, 1, "shot.png", nacelle.Image{MediaType: "image/png", Data: png}),
		upload(t, f.store, 1, "spec.pdf", nacelle.Image{MediaType: "application/pdf", Data: []byte("%PDF-1.7")}),
		upload(t, f.store, 1, "main.go", nacelle.Image{MediaType: "text/plain", Data: []byte("package main")}),
	}
	runID, err := f.runner.StartWithAttachments(context.Background(), 1, f.session.ID, "", ids)
	if err != nil {
		t.Fatal(err)
	}
	f.waitDone(t, runID)
	assertHydrated(t, backend.lastUser(), png)
	stored, _ := f.store.Messages(context.Background(), f.session.ID)
	if raw := string(stored[0].Payload); strings.Contains(raw, "PNG") || strings.Contains(raw, "package main") {
		t.Fatalf("the transcript holds file bytes: %s", raw)
	}
	session, _ := f.store.GetSession(context.Background(), 1, f.session.ID)
	if session.Title != "shot.png" {
		t.Errorf("title = %q", session.Title)
	}
}

func assertHydrated(t *testing.T, parts []nacelle.Part, png []byte) {
	t.Helper()
	if len(parts) != 3 {
		t.Fatalf("want 3 parts, got %+v", parts)
	}
	image, ok := parts[0].(nacelle.Image)
	if !ok || image.MediaType != "image/png" || !bytes.Equal(image.Data, png) {
		t.Errorf("image part: %#v", parts[0])
	}
	if doc, ok := parts[1].(nacelle.Document); !ok || doc.Name != "spec.pdf" {
		t.Errorf("pdf part: %#v", parts[1])
	}
	if text, ok := parts[2].(nacelle.Text); !ok || text.Text != "<file name=\"main.go\">\npackage main\n</file>" {
		t.Errorf("text part: %#v", parts[2])
	}
}

func TestGoogleGetsANoteInsteadOfAPDF(t *testing.T) {
	backend := &captureBackend{}
	f := newFixture(t, backend, nil, nil)
	session, _ := f.store.CreateSession(context.Background(), store.Session{ID: store.NewID(), UserID: 1, Channel: store.ChannelWeb, Backend: "google"})
	id := upload(t, f.store, 1, "spec.pdf", nacelle.Image{MediaType: "application/pdf", Data: []byte("%PDF-1.7")})
	runID, err := f.runner.StartWithAttachments(context.Background(), 1, session.ID, "summarise", []string{id})
	if err != nil {
		t.Fatal(err)
	}
	f.waitDone(t, runID)
	parts := backend.lastUser()
	note, ok := parts[0].(nacelle.Text)
	if !ok || !strings.Contains(note.Text, "cannot read PDFs") {
		t.Fatalf("want a PDF note, got %#v", parts)
	}
}

func TestAttachmentsAreValidated(t *testing.T) {
	f := newFixture(t, &captureBackend{}, nil, nil)
	foreign := upload(t, f.store, 2, "theirs.txt", nacelle.Image{MediaType: "text/plain", Data: []byte("x")})
	if _, err := f.runner.StartWithAttachments(context.Background(), 1, f.session.ID, "", nil); !errors.Is(err, runner.ErrEmptyMessage) {
		t.Fatalf("empty: %v", err)
	}
	if _, err := f.runner.StartWithAttachments(context.Background(), 1, f.session.ID, "hi", []string{foreign}); !errors.Is(err, runner.ErrUnknownAttachment) {
		t.Fatalf("foreign: %v", err)
	}
	many := make([]string, runner.MaxAttachments+1)
	if _, err := f.runner.StartWithAttachments(context.Background(), 1, f.session.ID, "hi", many); !errors.Is(err, runner.ErrTooManyAttachments) {
		t.Fatalf("too many: %v", err)
	}
}

func TestFileTextIsCapped(t *testing.T) {
	text := runner.FileText("big.log", bytes.Repeat([]byte("a"), 300<<10)).Text
	if len(text) > 201<<10 || !strings.Contains(text, "truncated at 200 KB") {
		t.Fatalf("not capped: %d", len(text))
	}
}
