package runner_test

import (
	"context"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

func TestGuestRunsDoNotHoldTheOwnersSlots(t *testing.T) {
	backend := &fakeBackend{block: make(chan struct{})}
	f := newFixture(t, backend, nil, nil)
	guest, err := f.store.CreateSession(context.Background(), store.Session{ID: store.NewID(), UserID: 1, Channel: store.ChannelGuest})
	if err != nil {
		t.Fatal(err)
	}
	var runs []string
	for range 2 {
		session := f.newSession(t, 1)
		runID, err := f.runner.Start(context.Background(), 1, session.ID, "hi")
		if err != nil {
			t.Fatal(err)
		}
		runs = append(runs, runID)
	}
	guestRun, err := f.runner.Start(context.Background(), 1, guest.ID, "hello from a colleague")
	if err != nil {
		t.Fatalf("a guest waited on the owner's slots: %v", err)
	}
	close(backend.block)
	for _, id := range append(runs, guestRun) {
		f.waitDone(t, id)
	}
}
