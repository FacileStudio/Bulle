package suite

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestAgendaUpdateKeepsUntouchedFieldsAndDuration(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	start := "2026-11-05T14:00:00+01:00"
	_, err := NewAgenda(NewGateway(w.tokens(), w.cfg, nil)).Update(context.Background(), 1, 4, 0, AgendaEventFields{StartAt: &start})
	if err != nil || len(w.writes) != 1 || w.writes[0].method != "PUT" || w.writes[0].path != "/api/events/4" {
		t.Fatalf("want one PUT to the event, got %+v, %v", w.writes, err)
	}
	var sent map[string]any
	_ = json.Unmarshal([]byte(w.writes[0].body), &sent)
	if sent["title"] != "standup" || sent["end_at"] != "2026-11-05T14:30:00+01:00" || sent["calendar_id"] != float64(3) {
		t.Fatalf("update must keep the title and calendar and shift the end by the old duration: %v", sent)
	}
	if _, leaked := sent["etag"]; leaked {
		t.Fatalf("read-only fields must not be sent back: %v", sent)
	}
}

func TestWriteWithoutGrantedScopeExplainsItself(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	w.readOnly = true
	gw := NewGateway(w.tokens(), w.cfg, nil)
	if _, err := NewAgenda(gw).Calendars(context.Background(), 1); err != nil {
		t.Fatalf("reads must keep working when writes are not granted: %v", err)
	}
	err := NewAgenda(gw).Delete(context.Background(), 1, 4)
	if !errors.Is(err, ErrWriteNotGranted) || !strings.Contains(err.Error(), "agenda.write") || len(w.writes) != 0 {
		t.Fatalf("want a write-not-granted error and no request, got %v, %d writes", err, len(w.writes))
	}
}

func TestReplyThreadsOnTheOriginal(t *testing.T) {
	original := courrierOriginal{MessageID: "<m2@x>", Subject: "Lunch", FromAddress: "ana@x", References: []string{"<m1@x>"}}
	got := threaded(sendRequest{Body: "yes"}, original)
	if got.Subject != "Re: Lunch" || got.To[0] != "ana@x" || got.InReplyTo != "<m2@x>" || strings.Join(got.References, " ") != "<m1@x> <m2@x>" {
		t.Fatalf("reply not threaded: %+v", got)
	}
	if again := threaded(sendRequest{Body: "ok"}, courrierOriginal{Subject: "RE: Lunch"}); again.Subject != "RE: Lunch" {
		t.Fatalf("an existing Re: must not be doubled: %q", again.Subject)
	}
}
