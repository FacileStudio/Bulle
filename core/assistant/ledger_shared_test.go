package assistant

import (
	"sync"
	"testing"
	"time"
)

func twoLedgers(t *testing.T) (*Ledger, *Ledger, FileState) {
	t.Helper()
	a, st := newTestLedger(t)
	b, err := NewLedger(st)
	if err != nil {
		t.Fatal(err)
	}
	return a, b, st
}

func TestLedgerStaleInstanceKeepsOtherWrites(t *testing.T) {
	a, b, st := twoLedgers(t)
	now := time.Now()
	one, two := Item{Source: SourceMail, ID: "1"}, Item{Source: SourceMail, ID: "2"}
	if err := a.MarkSurfaced([]Item{one}, now); err != nil {
		t.Fatal(err)
	}
	if err := b.MarkSurfaced([]Item{two}, now); err != nil {
		t.Fatal(err)
	}
	fresh, err := NewLedger(st)
	if err != nil {
		t.Fatal(err)
	}
	if fresh.ShouldSurface(one, now) || fresh.ShouldSurface(two, now) {
		t.Fatal("a mark was lost")
	}
	if b.ShouldSurface(one, now) {
		t.Fatal("stale reader missed the other process write")
	}
}

func TestLedgerTwoInstancesConcurrent(t *testing.T) {
	a, b, st := twoLedgers(t)
	now := time.Now()
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			l := a
			if i%2 == 1 {
				l = b
			}
			it := Item{Source: SourceMail, ID: string(rune('a' + i))}
			if err := l.MarkSurfaced([]Item{it}, now); err != nil {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()
	fresh, err := NewLedger(st)
	if err != nil || len(fresh.entries) != 20 {
		t.Fatalf("entries %d err %v", len(fresh.entries), err)
	}
}

func TestSnoozeExpiryResurfacesOnce(t *testing.T) {
	l, _ := newTestLedger(t)
	now := time.Date(2026, 10, 30, 10, 0, 0, 0, time.UTC)
	item := landingPage("todo")
	if err := l.MarkSurfaced([]Item{item}, now); err != nil {
		t.Fatal(err)
	}
	if err := l.Snooze(item.Key(), now, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	later := now.Add(2 * time.Hour)
	if !l.ShouldSurface(item, later) {
		t.Fatal("expired snooze must surface")
	}
	if err := l.MarkSurfaced([]Item{item}, later); err != nil {
		t.Fatal(err)
	}
	if l.ShouldSurface(item, later.Add(time.Hour)) {
		t.Fatal("must surface only once")
	}
}

func TestNoiseOnlyMarksMailAndDropsSender(t *testing.T) {
	l, st := newTestLedger(t)
	now := time.Now()
	if err := RecordMailSenders(st, map[string]string{"mail:t1": "Promo@Shop.test"}); err != nil {
		t.Fatal(err)
	}
	task := landingPage("todo")
	if err := l.MarkNoiseAt(task.Key(), now); err != nil {
		t.Fatal(err)
	}
	if !l.ShouldSurface(task, now) {
		t.Fatal("task became noise")
	}
	if err := l.MarkNoiseAt("mail:t1", now); err != nil {
		t.Fatal(err)
	}
	drop := LoadDropSenders(st)
	if len(drop) != 1 || drop[0] != "promo@shop.test" {
		t.Fatalf("drop list = %v", drop)
	}
}
