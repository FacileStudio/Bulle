package scheduler_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/scheduler"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

type ideaCalls struct {
	mu     sync.Mutex
	spaces []string
}

func (c *ideaCalls) GenerateIdeas(_ context.Context, _ int64, spaceID, trigger string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if trigger == store.BatchDaily {
		c.spaces = append(c.spaces, spaceID)
	}
	return "batch", nil
}

func (c *ideaCalls) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.spaces)
}

func TestDailyIdeasFireAtTheLocalHourOncePerDay(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	paris, _ := st.CreateSpace(ctx, store.Space{UserID: 1, Name: "Paris"})
	off, _ := st.CreateSpace(ctx, store.Space{UserID: 1, Name: "Off"})
	settings := store.DefaultIdeaSettings(1, paris.ID)
	settings.Timezone = "Europe/Paris"
	if _, err := st.SaveIdeaSettings(ctx, settings); err != nil {
		t.Fatal(err)
	}
	disabled := store.DefaultIdeaSettings(1, off.ID)
	disabled.Enabled, disabled.Hour = false, 7
	if _, err := st.SaveIdeaSettings(ctx, disabled); err != nil {
		t.Fatal(err)
	}
	calls := &ideaCalls{}
	sched := scheduler.New(scheduler.Config{Store: st, Runner: newMockRunner(), Ideas: calls})

	sched.Tick(ctx, time.Date(2026, 11, 6, 8, 0, 0, 0, time.UTC))
	if calls.count() != 0 {
		t.Fatal("08:00 UTC is 09:00 in Paris; the batch fired at the wrong hour")
	}
	sched.Tick(ctx, time.Date(2026, 11, 6, 7, 0, 0, 0, time.UTC))
	sched.Tick(ctx, time.Date(2026, 11, 6, 7, 30, 0, 0, time.UTC))
	if calls.count() != 1 || calls.spaces[0] != paris.ID {
		t.Fatalf("want one daily batch for Paris at 08:00 local, got %v", calls.spaces)
	}
	sched.Tick(ctx, time.Date(2026, 11, 7, 7, 5, 0, 0, time.UTC))
	if calls.count() != 2 {
		t.Fatalf("the next local day must fire again, got %d", calls.count())
	}
}

func TestTickExpiresIdeasPastTheirDate(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	space, _ := st.EnsureDefaultSpace(ctx, 1)
	now := time.Now()
	idea, _ := st.CreateIdea(ctx, store.Idea{UserID: 1, SpaceID: space.ID, BatchID: "b", Kind: "task", Title: "t",
		Why: "w", Prompt: "p", Fingerprint: "f", ExpiresAt: now.Add(-time.Minute)})
	scheduler.New(scheduler.Config{Store: st, Runner: newMockRunner()}).Tick(ctx, now)
	if got, _ := st.GetIdea(ctx, 1, idea.ID); got.Status != store.IdeaExpired {
		t.Fatalf("idea past expires_at is %s", got.Status)
	}
}
