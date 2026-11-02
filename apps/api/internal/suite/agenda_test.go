package suite

import (
	"testing"
	"time"
)

func TestAgendaWindowUsesTheUsersDay(t *testing.T) {
	now := time.Date(2026, 11, 2, 23, 30, 0, 0, time.UTC)
	agenda := Agenda{now: func() time.Time { return now }}
	from, to, err := agenda.window(AgendaQuery{Timezone: "Europe/Paris"})
	if err != nil {
		t.Fatal(err)
	}
	if got := from.Format(time.RFC3339); got != "2026-11-03T00:00:00+01:00" {
		t.Errorf("from = %s", got)
	}
	if got := to.Format(time.RFC3339); got != "2026-11-10T00:00:00+01:00" {
		t.Errorf("to = %s", got)
	}
}

func TestAgendaWindowForms(t *testing.T) {
	agenda := Agenda{now: func() time.Time { return time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC) }}
	cases := []struct{ from, to, wantFrom, wantTo string }{
		{"tomorrow", "", "2026-03-29T00:00:00Z", "2026-04-05T00:00:00Z"},
		{"2026-04-01", "2026-04-02", "2026-04-01T00:00:00Z", "2026-04-02T00:00:00Z"},
		{"2026-04-01T09:00:00+02:00", "", "2026-04-01T09:00:00+02:00", "2026-04-08T09:00:00+02:00"},
	}
	for _, c := range cases {
		from, to, err := agenda.window(AgendaQuery{From: c.from, To: c.to})
		if err != nil || from.Format(time.RFC3339) != c.wantFrom || to.Format(time.RFC3339) != c.wantTo {
			t.Errorf("%+v: got %s %s, %v", c, from.Format(time.RFC3339), to.Format(time.RFC3339), err)
		}
	}
	if _, _, err := agenda.window(AgendaQuery{From: "soon"}); err == nil {
		t.Error("a non-date must be refused rather than ignored")
	}
	if _, _, err := agenda.window(AgendaQuery{Timezone: "Mars/Base"}); err == nil {
		t.Error("an unknown timezone must be refused")
	}
}
