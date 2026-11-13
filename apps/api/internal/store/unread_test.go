package store_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

func unreadOf(t *testing.T, s store.Store, owner int64, id string) bool {
	t.Helper()
	got, err := s.GetSession(context.Background(), owner, id)
	if err != nil {
		t.Fatal(err)
	}
	return got.Unread()
}

func openSession(t *testing.T, s store.Store, owner int64, channel string) string {
	t.Helper()
	created, err := s.CreateSession(context.Background(), store.Session{ID: store.NewID(), UserID: owner, Channel: channel, Title: channel})
	if err != nil {
		t.Fatal(err)
	}
	return created.ID
}

func exerciseUnread(t *testing.T, s store.Store, owner, stranger int64) {
	t.Helper()
	web, cron, matrix := openSession(t, s, owner, store.ChannelWeb), openSession(t, s, owner, store.ChannelCron), openSession(t, s, owner, store.ChannelMatrix)
	if unreadOf(t, s, owner, web) || !unreadOf(t, s, owner, cron) {
		t.Fatal("a new conversation is unread only when a job started it")
	}
	markRead(t, s, owner, stranger, cron)
	time.Sleep(2 * time.Millisecond)
	reply := []json.RawMessage{json.RawMessage(`{"role":"assistant"}`)}
	for _, id := range []string{web, cron, matrix} {
		if err := s.ReplaceMessages(context.Background(), id, reply); err != nil {
			t.Fatal(err)
		}
	}
	if !unreadOf(t, s, owner, web) || !unreadOf(t, s, owner, cron) {
		t.Fatal("a new message does not make the conversation unread")
	}
	if unreadOf(t, s, owner, matrix) {
		t.Fatal("a Matrix thread shows unread in the web app")
	}
}

func markRead(t *testing.T, s store.Store, owner, stranger int64, id string) {
	t.Helper()
	ctx := context.Background()
	if err := s.MarkSessionRead(ctx, stranger, id); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("stranger marked read: %v", err)
	}
	if err := s.MarkSessionRead(ctx, owner, "missing"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("missing session marked read: %v", err)
	}
	if err := s.MarkSessionRead(ctx, owner, id); err != nil {
		t.Fatal(err)
	}
	if unreadOf(t, s, owner, id) {
		t.Fatal("opening the conversation leaves it unread")
	}
}
