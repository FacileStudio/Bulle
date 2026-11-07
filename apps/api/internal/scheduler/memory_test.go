package scheduler_test

import (
	"context"
	"encoding/json"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/scheduler"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

type memoryCalls struct {
	st      *store.Memory
	without int64

	mu       sync.Mutex
	sessions []string
	asked    map[int64]int
}

func (c *memoryCalls) MemoryAvailable(_ context.Context, userID int64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.asked[userID]++
	return userID != c.without
}

func (c *memoryCalls) MemoryPass(ctx context.Context, userID int64, sessionID string) error {
	messages, _ := c.st.Messages(ctx, sessionID)
	c.mu.Lock()
	c.sessions = append(c.sessions, sessionID)
	c.mu.Unlock()
	return c.st.MarkMemorySeen(ctx, sessionID, len(messages))
}

func (c *memoryCalls) passed() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return slices.Clone(c.sessions)
}

func chat(t *testing.T, st *store.Memory, userID int64, channel string, messages int) store.Session {
	t.Helper()
	ctx := context.Background()
	session, err := st.CreateSession(ctx, store.Session{ID: store.NewID(), UserID: userID, Channel: channel})
	if err != nil {
		t.Fatal(err)
	}
	payloads := make([]json.RawMessage, messages)
	for i := range payloads {
		payloads[i] = json.RawMessage(`{"role":"user","parts":[{"type":"text","text":"hello"}]}`)
	}
	if err := st.ReplaceMessages(ctx, session.ID, payloads); err != nil {
		t.Fatal(err)
	}
	return session
}

func TestMemoryPassesIdleChatsOnceAndMarksThemSeen(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	web := chat(t, st, 1, store.ChannelWeb, 2)
	matrix := chat(t, st, 1, store.ChannelMatrix, 1)
	chat(t, st, 1, store.ChannelCron, 2)
	chat(t, st, 1, store.ChannelIdeas, 2)
	chat(t, st, 1, store.ChannelWeb, 0)
	chat(t, st, 2, store.ChannelWeb, 3)
	calls := &memoryCalls{st: st, without: 2, asked: map[int64]int{}}
	sched := scheduler.New(scheduler.Config{Store: st, Runner: newMockRunner(), Memory: calls})

	sched.Tick(ctx, time.Now().Add(5*time.Minute))
	if got := calls.passed(); len(got) != 0 {
		t.Fatalf("a chat quiet for five minutes got a memory pass: %v", got)
	}
	sched.Tick(ctx, time.Now().Add(time.Hour))
	got := calls.passed()
	if len(got) != 2 || !slices.Contains(got, web.ID) || !slices.Contains(got, matrix.ID) {
		t.Fatalf("want one pass for the idle web and matrix chats only, got %v", got)
	}
	if calls.asked[2] != 1 {
		t.Fatalf("want the user without Mycelium asked about once per tick, got %d", calls.asked[2])
	}
	seen, _ := st.GetSession(ctx, 1, web.ID)
	if seen.MemorySeenSeq != 2 {
		t.Fatalf("the chat is not marked seen: %d", seen.MemorySeenSeq)
	}
	sched.Tick(ctx, time.Now().Add(2*time.Hour))
	if again := calls.passed(); len(again) != 2 {
		t.Fatalf("a chat with nothing new got a second pass: %v", again)
	}
}

func TestMemoryPassesAreCappedPerTick(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	for range scheduler.MemoryPassesPerTick + 3 {
		chat(t, st, 1, store.ChannelWeb, 1)
	}
	calls := &memoryCalls{st: st, asked: map[int64]int{}}
	sched := scheduler.New(scheduler.Config{Store: st, Runner: newMockRunner(), Memory: calls})
	sched.Tick(ctx, time.Now().Add(time.Hour))
	if got := len(calls.passed()); got != scheduler.MemoryPassesPerTick {
		t.Fatalf("want %d passes in one tick, got %d", scheduler.MemoryPassesPerTick, got)
	}
	sched.Tick(ctx, time.Now().Add(time.Hour))
	if got := len(calls.passed()); got != scheduler.MemoryPassesPerTick+3 {
		t.Fatalf("want the rest on the next tick, got %d", got)
	}
}
