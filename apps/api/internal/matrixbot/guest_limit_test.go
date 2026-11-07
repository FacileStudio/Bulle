package matrixbot

import (
	"testing"
	"time"
)

func TestGuestLimiterPerSender(t *testing.T) {
	var l guestLimiter
	now := time.Date(2026, 11, 7, 9, 0, 0, 0, time.UTC)
	for i := range guestBurst {
		if !l.allow(1, "@ann:x", now.Add(time.Duration(i)*time.Second)) {
			t.Fatalf("message %d refused", i)
		}
	}
	if l.allow(1, "@ann:x", now.Add(time.Minute)) {
		t.Fatal("a guest went past the burst")
	}
	if !l.allow(1, "@bob:x", now.Add(time.Minute)) || !l.allow(2, "@ann:x", now.Add(time.Minute)) {
		t.Fatal("the limit leaked to another sender or owner")
	}
	if !l.allow(1, "@ann:x", now.Add(guestWindow+time.Minute)) {
		t.Fatal("the window never reopened")
	}
}
