package runner

import (
	"context"
	"sync"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

type countingFiles struct {
	store.Attachments
	gets int
}

func (c *countingFiles) GetAttachment(context.Context, int64, string) (store.Attachment, error) {
	c.gets++
	return store.Attachment{Name: "notes.txt", MediaType: "text/plain", Data: []byte("hello")}, nil
}

func TestHydratorLoadsEachAttachmentOncePerRun(t *testing.T) {
	files := &countingFiles{}
	h := hydrator{files: files, userID: 1, pdf: true, loaded: &sync.Map{}}
	ref := AttachmentPart(AttachmentRef{ID: "a1", Name: "notes.txt", MediaType: "text/plain"})
	messages := []nacelle.Message{{Role: nacelle.RoleUser, Parts: []nacelle.Part{ref}}}
	for range 3 {
		out, err := h.hydrate(context.Background(), messages)
		if err != nil {
			t.Fatal(err)
		}
		if text, ok := out[0].Parts[0].(nacelle.Text); !ok || text.Text == ref.Text {
			t.Fatalf("the reference was not hydrated: %#v", out[0].Parts[0])
		}
	}
	if files.gets != 1 {
		t.Errorf("GetAttachment ran %d times, want once per run", files.gets)
	}
}
