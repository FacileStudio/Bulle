package store_test

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

func TestMemoryChats(t *testing.T) {
	exerciseChats(t, store.NewMemory(), 1, 2)
}

func TestPostgresChats(t *testing.T) {
	if testDB == nil {
		t.Skip("no TEST_DATABASE_URL")
	}
	if err := testDB.Exec("TRUNCATE users RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatal(err)
	}
	for _, id := range []int{1, 2} {
		err := testDB.Exec("INSERT INTO users (id, email) VALUES (?, ?)", id, fmt.Sprintf("c%d@facile.studio", id)).Error
		if err != nil {
			t.Fatal(err)
		}
	}
	exerciseChats(t, store.NewPostgres(testDB), 1, 2)
}

func text(role, body string) json.RawMessage {
	data, _ := json.Marshal(map[string]any{"role": role, "parts": []map[string]string{{"type": "text", "text": body}}})
	return data
}

func toolOutput(body string) json.RawMessage {
	data, _ := json.Marshal(map[string]any{"role": "user", "parts": []map[string]string{{"type": "tool_result", "id": "c1", "name": "courrier_list_inbox", "result": body}}})
	return data
}

type chatSetup struct {
	s         store.Store
	userID    int64
	space     string
	sessionOf map[string]string
}

func (c chatSetup) add(t *testing.T, name, channel, title string, messages ...json.RawMessage) {
	t.Helper()
	ctx := context.Background()
	session, err := c.s.CreateSession(ctx, store.Session{ID: store.NewID(), UserID: c.userID, SpaceID: c.space, Channel: channel, Title: title})
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if err := c.s.ReplaceMessages(ctx, session.ID, messages); err != nil {
		t.Fatal(err)
	}
	c.sessionOf[name] = session.ID
}

func exerciseChats(t *testing.T, s store.Store, owner, stranger int64) {
	t.Helper()
	ctx := context.Background()
	home, err := s.EnsureDefaultSpace(ctx, owner)
	if err != nil {
		t.Fatal(err)
	}
	work, err := s.CreateSpace(ctx, store.Space{UserID: owner, Name: "Work"})
	if err != nil {
		t.Fatal(err)
	}
	theirs, err := s.EnsureDefaultSpace(ctx, stranger)
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]string{}
	planning := []json.RawMessage{text("user", "Can we move the quarterly PLANNING meeting?"), text("assistant", "Sure.")}
	mine := chatSetup{s: s, userID: owner, space: home.ID, sessionOf: ids}
	mine.add(t, "web", store.ChannelWeb, "Budget review", planning...)
	mine.add(t, "matrix", store.ChannelMatrix, "", text("assistant", "Your planning notes are ready."))
	mine.add(t, "tool", store.ChannelWeb, "Inbox", text("user", "check mail"), toolOutput("planning offsite"), text("assistant", "Nothing new."))
	for _, channel := range []string{store.ChannelCron, store.ChannelIdeas, store.ChannelMemory, store.ChannelGuest} {
		mine.add(t, channel, channel, "planning", planning...)
	}
	chatSetup{s: s, userID: owner, space: work.ID, sessionOf: ids}.add(t, "work", store.ChannelWeb, "", planning...)
	chatSetup{s: s, userID: stranger, space: theirs.ID, sessionOf: ids}.add(t, "stranger", store.ChannelWeb, "", planning...)

	assertSearch(t, s, owner, home.ID, "planning", ids["web"], ids["matrix"])
	assertSearch(t, s, owner, work.ID, "planning", ids["work"])
	assertSearch(t, s, stranger, home.ID, "planning")
	assertSearch(t, s, owner, home.ID, "%_", []string{}...)
	hits, _ := s.SearchSessions(ctx, owner, home.ID, "quarterly planning", 20)
	if len(hits) != 1 || !strings.Contains(hits[0].Snippet, "quarterly PLANNING meeting") {
		t.Fatalf("want the snippet around the match, got %+v", hits)
	}
	hits, _ = s.SearchSessions(ctx, owner, home.ID, "BUDGET", 20)
	if len(hits) != 1 || hits[0].SessionID != ids["web"] || hits[0].Title != "Budget review" || hits[0].Snippet != "" {
		t.Fatalf("want a title match with no snippet, got %+v", hits)
	}
	if hits, _ := s.SearchSessions(ctx, owner, home.ID, "planning", 1); len(hits) != 1 {
		t.Fatalf("the limit is ignored: %d hits", len(hits))
	}
	listed, _ := s.ListSessions(ctx, owner, "")
	for _, session := range listed {
		if store.Hidden(session.Channel) {
			t.Fatalf("a %s session shows in the session list", session.Channel)
		}
	}
	exerciseMemoryPassQueue(t, s, owner, ids)
}

func assertSearch(t *testing.T, s store.Store, userID int64, spaceID, query string, want ...string) {
	t.Helper()
	hits, err := s.SearchSessions(context.Background(), userID, spaceID, query, 20)
	if err != nil {
		t.Fatal(err)
	}
	got := make([]string, len(hits))
	for i, hit := range hits {
		got[i] = hit.SessionID
	}
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("search %q in %s: want %v, got %v", query, spaceID, want, got)
	}
}

func exerciseMemoryPassQueue(t *testing.T, s store.Store, owner int64, ids map[string]string) {
	t.Helper()
	ctx := context.Background()
	if due, _ := s.ListSessionsForMemoryPass(ctx, time.Now().Add(-time.Minute)); len(due) != 0 {
		t.Fatalf("a chat that just changed is due for a memory pass: %d", len(due))
	}
	queued := func() []string {
		due, err := s.ListSessionsForMemoryPass(ctx, time.Now().Add(time.Hour))
		if err != nil {
			t.Fatal(err)
		}
		out := []string{}
		for _, session := range due {
			out = append(out, session.ID)
		}
		slices.Sort(out)
		return out
	}
	want := []string{ids["web"], ids["matrix"], ids["tool"], ids["work"], ids["stranger"]}
	slices.Sort(want)
	if got := queued(); !slices.Equal(got, want) {
		t.Fatalf("want only web and matrix chats queued, got %v want %v", got, want)
	}
	if err := s.MarkMemorySeen(ctx, ids["web"], 2); err != nil {
		t.Fatal(err)
	}
	if err := s.MarkMemorySeen(ctx, ids["web"], 1); err != nil {
		t.Fatal(err)
	}
	session, _ := s.GetSession(ctx, owner, ids["web"])
	if session.MemorySeenSeq != 2 {
		t.Fatalf("the seen mark went back: %d", session.MemorySeenSeq)
	}
	if session.Tainted {
		t.Fatal("a new chat starts tainted")
	}
	if err := s.MarkTainted(ctx, ids["web"]); err != nil {
		t.Fatal(err)
	}
	if session, _ = s.GetSession(ctx, owner, ids["web"]); !session.Tainted {
		t.Fatal("the taint mark was not kept")
	}
	if got := queued(); slices.Contains(got, ids["web"]) {
		t.Fatal("a chat with nothing new is still queued")
	}
	if err := s.ReplaceMessages(ctx, ids["web"], append([]json.RawMessage{text("user", "a"), text("assistant", "b")}, text("user", "c"))); err != nil {
		t.Fatal(err)
	}
	if got := queued(); !slices.Contains(got, ids["web"]) {
		t.Fatal("a chat with a new message is not queued again")
	}
}
