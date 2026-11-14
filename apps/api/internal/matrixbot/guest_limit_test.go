package matrixbot

import (
	"strconv"
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

func TestGuestLimiterIsASlidingWindowWithBoundedKeys(t *testing.T) {
	var l guestLimiter
	now := time.Date(2026, 11, 7, 9, 0, 0, 0, time.UTC)
	allowed := 0
	for minute := range 10 {
		for range 3 {
			if l.allow(1, "@ann:x", now.Add(time.Duration(minute)*time.Minute)) {
				allowed++
			}
		}
	}
	if allowed != guestBurst {
		t.Fatalf("%d messages in one window, want %d", allowed, guestBurst)
	}
	for i := range guestKeys {
		l.allow(2, strconv.Itoa(i), now)
	}
	if l.allow(3, "@new:x", now.Add(time.Minute)) {
		t.Fatal("a new sender got in while every remembered pair was active")
	}
	if !l.allow(3, "@new:x", now.Add(guestWindow+time.Minute)) || len(l.seen) > guestKeys {
		t.Fatalf("quiet pairs were not forgotten: %d remembered", len(l.seen))
	}
}
