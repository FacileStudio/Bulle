package store_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

func exerciseAttachments(t *testing.T, s store.Store, owner, stranger int64) {
	t.Helper()
	ctx := context.Background()
	data := []byte("\x89PNG body")
	created, err := s.CreateAttachment(ctx, store.Attachment{ID: store.NewID(), UserID: owner, Name: "a.png", MediaType: "image/png", Size: int64(len(data)), Data: data})
	if err != nil || created.CreatedAt.IsZero() || len(created.Data) != 0 {
		t.Fatalf("create: %v %+v", err, created)
	}
	got, err := s.GetAttachment(ctx, owner, created.ID)
	if err != nil || !bytes.Equal(got.Data, data) || got.MediaType != "image/png" {
		t.Fatalf("get: %v %+v", err, got)
	}
	if _, err := s.GetAttachment(ctx, stranger, created.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("stranger get: %v", err)
	}
	info, err := s.AttachmentInfo(ctx, owner, []string{"missing", created.ID})
	if err != nil || len(info) != 1 || info[0].Name != "a.png" || len(info[0].Data) != 0 {
		t.Fatalf("info: %v %+v", err, info)
	}
	if other, _ := s.AttachmentInfo(ctx, stranger, []string{created.ID}); len(other) != 0 {
		t.Fatalf("stranger info: %+v", other)
	}
	exerciseMyceliumSpaces(t, s, owner)
}

func exerciseMyceliumSpaces(t *testing.T, s store.Store, owner int64) {
	t.Helper()
	ctx := context.Background()
	if _, err := s.MyceliumSpace(ctx, owner); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("unset space: %v", err)
	}
	for _, id := range []string{"sp-1", "sp-2"} {
		if err := s.SaveMyceliumSpace(ctx, owner, id); err != nil {
			t.Fatal(err)
		}
	}
	if id, err := s.MyceliumSpace(ctx, owner); err != nil || id != "sp-2" {
		t.Fatalf("space: %q %v", id, err)
	}
}
