package suite

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	agendaCalendarsPath = "/api/calendars/"
	agendaEventsPath    = "/api/calendars/%d/events"
	agendaDefaultDays   = 7
)

var agendaEventKeys = []string{
	"id", "calendar_id", "title", "description", "location", "start_at", "end_at",
	"is_all_day", "recurrence_rule", "status", "conference_url", "attendees",
}

type AgendaCalendar struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	IsPersonal bool   `json:"is_personal"`
	Role       string `json:"role"`
}

type AgendaQuery struct {
	Calendar int64
	From     string
	To       string
	Timezone string
}

type Agenda struct {
	gw  *Gateway
	now func() time.Time
}

func NewAgenda(gw *Gateway) Agenda { return Agenda{gw: gw, now: time.Now} }

func (a Agenda) Calendars(ctx context.Context, userID int64) ([]AgendaCalendar, error) {
	raw, err := a.gw.Get(ctx, userID, AppAgenda, agendaCalendarsPath, nil)
	if err != nil {
		return nil, err
	}
	var calendars []AgendaCalendar
	if err := decodeInto(raw, &calendars); err != nil {
		return nil, fmt.Errorf("agenda: reading calendars: %w", err)
	}
	return calendars, nil
}

func (a Agenda) Events(ctx context.Context, userID int64, q AgendaQuery) ([]byte, error) {
	from, to, err := a.window(q)
	if err != nil {
		return nil, err
	}
	calendars, err := a.target(ctx, userID, q.Calendar)
	if err != nil {
		return nil, err
	}
	query := url.Values{"from": {from.Format(time.RFC3339)}, "to": {to.Format(time.RFC3339)}}
	if q.Timezone != "" {
		query.Set("tz", q.Timezone)
	}
	var merged []byte
	for _, id := range calendars {
		raw, err := a.gw.Get(ctx, userID, AppAgenda, fmt.Sprintf(agendaEventsPath, id), query)
		if err != nil {
			return nil, err
		}
		merged, err = appendJSONList(merged, raw)
		if err != nil {
			return nil, err
		}
	}
	if merged == nil {
		return []byte("[]"), nil
	}
	return pick(merged, agendaEventKeys...)
}

func (a Agenda) target(ctx context.Context, userID, requested int64) ([]int64, error) {
	if requested > 0 {
		return []int64{requested}, nil
	}
	calendars, err := a.Calendars(ctx, userID)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(calendars))
	for _, calendar := range calendars {
		ids = append(ids, calendar.ID)
	}
	return ids, nil
}

func (a Agenda) window(q AgendaQuery) (time.Time, time.Time, error) {
	location := time.UTC
	if q.Timezone != "" {
		loaded, err := time.LoadLocation(q.Timezone)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("%q is not an IANA timezone", q.Timezone)
		}
		location = loaded
	}
	now := a.now().In(location)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	from, err := parseWhen(q.From, today, location)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if q.From == "" {
		from = today
	}
	to, err := parseWhen(q.To, today, location)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if q.To == "" {
		to = from.AddDate(0, 0, agendaDefaultDays)
	}
	return from, to, nil
}

func parseWhen(value string, today time.Time, location *time.Location) (time.Time, error) {
	value = strings.TrimSpace(value)
	switch strings.ToLower(value) {
	case "", "today":
		return today, nil
	case "tomorrow":
		return today.AddDate(0, 0, 1), nil
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}
	if parsed, err := time.ParseInLocation("2006-01-02", value, location); err == nil {
		return parsed, nil
	}
	return time.Time{}, fmt.Errorf("%q is not a date: use today, tomorrow, YYYY-MM-DD or RFC 3339", value)
}
