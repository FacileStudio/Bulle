package live_test

import (
	"context"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/live"
	"github.com/FacileStudio/Bulle/apps/api/internal/push"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

func TestStorePublishesSessionChangesToTheOwnerOnly(t *testing.T) {
	ctx := context.Background()
	hub := push.NewHub(push.HubConfig{})
	mine, stop := hub.Subscribe(1)
	defer stop()
	other, stopOther := hub.Subscribe(2)
	defer stopOther()
	st := live.Wrap(store.NewMemory(), hub)

	created, err := st.CreateSession(ctx, store.Session{ID: store.NewID(), UserID: 1, Title: "hi"})
	if err != nil {
		t.Fatal(err)
	}
	if err := st.DeleteSession(ctx, 2, created.ID); err == nil {
		t.Fatal("a stranger deleted the session")
	}
	if err := st.DeleteSession(ctx, 1, created.ID); err != nil {
		t.Fatal(err)
	}
	if len(mine) != 2 || len(other) != 0 {
		t.Fatalf("owner got %d events, stranger %d", len(mine), len(other))
	}
	first, second := <-mine, <-mine
	if first.Type != live.SessionCreated || second.Type != live.SessionDeleted {
		t.Fatalf("events %s, %s", first.Type, second.Type)
	}
	if string(second.Data) != `{"session_id":"`+created.ID+`"}` {
		t.Fatalf("session.deleted data %s", second.Data)
	}
	if live.Wrap(st, nil) != st {
		t.Fatal("a nil publisher should leave the store as is")
	}
}

func TestMarkReadTellsTheOwnerOnly(t *testing.T) {
	ctx := context.Background()
	hub := push.NewHub(push.HubConfig{})
	mine, stop := hub.Subscribe(1)
	defer stop()
	other, stopOther := hub.Subscribe(2)
	defer stopOther()
	st := live.Wrap(store.NewMemory(), hub)

	job, err := st.CreateSession(ctx, store.Session{ID: store.NewID(), UserID: 1, Channel: store.ChannelCron})
	if err != nil {
		t.Fatal(err)
	}
	if created := <-mine; !strings.Contains(string(created.Data), `"unread":true`) {
		t.Fatalf("a job conversation is announced read: %s", created.Data)
	}
	if err := st.MarkSessionRead(ctx, 2, job.ID); err == nil {
		t.Fatal("a stranger marked the session read")
	}
	if err := st.MarkSessionRead(ctx, 1, job.ID); err != nil {
		t.Fatal(err)
	}
	if read := <-mine; read.Type != live.SessionRead || len(other) != 0 {
		t.Fatalf("mark read sent %s, stranger got %d", read.Type, len(other))
	}
}
