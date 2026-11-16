package ideas

import (
	"testing"
	"time"
)

func TestARunLimiterCountsPerKeyInsideTheWindow(t *testing.T) {
	var l runLimiter
	start := time.Now()
	for i := 0; i < AppRunsPerHour; i++ {
		if !l.allow("agenda\x001", start) {
			t.Fatalf("run %d refused", i+1)
		}
	}
	if l.allow("agenda\x001", start.Add(time.Minute)) {
		t.Fatal("a run past the cap inside the hour was allowed")
	}
	if !l.allow("glouton\x001", start) || !l.allow("agenda\x002", start) {
		t.Fatal("another app or user shares the cap")
	}
	if !l.allow("agenda\x001", start.Add(runWindow+time.Second)) {
		t.Fatal("the cap did not reopen after the window")
	}
}
