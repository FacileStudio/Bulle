package push

import "testing"

func TestLiveCapsStreamsPerUserAndCloses(t *testing.T) {
	l := NewLive()
	first, _ := l.Subscribe(1)
	for range MaxStreams - 1 {
		l.Subscribe(1)
	}
	newest, stop := l.Subscribe(1)
	if _, open := <-first; open {
		t.Fatal("the oldest stream stays open past the cap")
	}
	other, _ := l.Subscribe(2)
	stop()
	l.Close()
	if _, open := <-other; open {
		t.Fatal("Close left a stream open")
	}
	if late, _ := l.Subscribe(1); late != nil {
		if _, open := <-late; open {
			t.Fatal("a stream opened after Close")
		}
	}
	select {
	case _, open := <-newest:
		if open {
			t.Fatal("an unsubscribed stream got an event")
		}
	default:
	}
}
