package suite

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const agendaEventPath = "/api/events/%d"

var agendaWritableKeys = []string{
	"calendar_id", "title", "description", "location", "start_at", "end_at", "is_all_day",
	"recurrence_rule", "status", "conference_url", "conference_provider", "timezone", "color",
	"attendees", "reminders",
}

type AgendaAttendee struct {
	Email string `json:"email" jsonschema:"required"`
	Name  string `json:"name,omitempty"`
}

type AgendaEventFields struct {
	Title          *string          `json:"title,omitempty" jsonschema:"description=Event title"`
	Description    *string          `json:"description,omitempty"`
	Location       *string          `json:"location,omitempty"`
	StartAt        *string          `json:"start_at,omitempty" jsonschema:"description=Start: RFC 3339 such as 2026-11-04T14:00:00+01:00, or YYYY-MM-DD for an all-day event"`
	EndAt          *string          `json:"end_at,omitempty" jsonschema:"description=End, same formats; defaults to one hour after the start (one day for all-day)"`
	IsAllDay       *bool            `json:"is_all_day,omitempty"`
	RecurrenceRule *string          `json:"recurrence_rule,omitempty" jsonschema:"description=RFC 5545 RRULE such as FREQ=WEEKLY;BYDAY=MO; empty clears it"`
	Status         *string          `json:"status,omitempty" jsonschema:"enum=confirmed,enum=tentative,enum=cancelled"`
	ConferenceURL  *string          `json:"conference_url,omitempty"`
	Timezone       *string          `json:"timezone,omitempty" jsonschema:"description=IANA timezone the event is anchored in, e.g. Europe/Paris"`
	Attendees      []AgendaAttendee `json:"attendees,omitempty" jsonschema:"description=Replaces the whole attendee list"`
}

func (a Agenda) Event(ctx context.Context, userID, eventID int64) ([]byte, error) {
	raw, err := a.gw.Get(ctx, userID, AppAgenda, fmt.Sprintf(agendaEventPath, eventID), nil)
	if err != nil {
		return nil, err
	}
	return pickOne(raw, agendaEventKeys...)
}

func (a Agenda) Create(ctx context.Context, userID, calendar int64, fields AgendaEventFields) ([]byte, error) {
	if fields.Title == nil || *fields.Title == "" || fields.StartAt == nil {
		return nil, errors.New("an event needs a title and a start")
	}
	id, err := a.writableCalendar(ctx, userID, calendar)
	if err != nil {
		return nil, err
	}
	body, err := a.payload(fields, nil)
	if err != nil {
		return nil, err
	}
	raw, err := a.gw.Do(ctx, userID, Request{App: AppAgenda, Method: http.MethodPost, Path: fmt.Sprintf(agendaEventsPath, id), Body: body})
	if err != nil {
		return nil, err
	}
	return pickOne(raw, agendaEventKeys...)
}

func (a Agenda) Update(ctx context.Context, userID, eventID, calendar int64, fields AgendaEventFields) ([]byte, error) {
	path := fmt.Sprintf(agendaEventPath, eventID)
	current, err := a.gw.Get(ctx, userID, AppAgenda, path, nil)
	if err != nil {
		return nil, err
	}
	stored, err := pickMap(current, agendaWritableKeys...)
	if err != nil {
		return nil, fmt.Errorf("agenda: reading event %d: %w", eventID, err)
	}
	body, err := a.payload(fields, stored)
	if err != nil {
		return nil, err
	}
	if calendar > 0 {
		body["calendar_id"] = calendar
	}
	raw, err := a.gw.Do(ctx, userID, Request{App: AppAgenda, Method: http.MethodPut, Path: path, Body: body})
	if err != nil {
		return nil, err
	}
	return pickOne(raw, agendaEventKeys...)
}

func (a Agenda) Delete(ctx context.Context, userID, eventID int64) error {
	_, err := a.gw.Do(ctx, userID, Request{App: AppAgenda, Method: http.MethodDelete, Path: fmt.Sprintf(agendaEventPath, eventID), Body: nil})
	return err
}

func (a Agenda) writableCalendar(ctx context.Context, userID, requested int64) (int64, error) {
	if requested > 0 {
		return requested, nil
	}
	calendars, err := a.Calendars(ctx, userID)
	if err != nil {
		return 0, err
	}
	for _, calendar := range calendars {
		if calendar.IsPersonal {
			return calendar.ID, nil
		}
	}
	if len(calendars) == 0 {
		return 0, errors.New("the user has no calendar in Agenda")
	}
	return calendars[0].ID, nil
}

func (a Agenda) payload(fields AgendaEventFields, stored map[string]any) (map[string]any, error) {
	body := map[string]any{}
	for key, value := range stored {
		body[key] = value
	}
	patch, err := toMap(fields)
	if err != nil {
		return nil, err
	}
	for key, value := range patch {
		body[key] = value
	}
	return body, a.schedule(body, fields, stored)
}

func (a Agenda) schedule(body map[string]any, fields AgendaEventFields, stored map[string]any) error {
	location, err := zone(fields.Timezone, stored)
	if err != nil {
		return err
	}
	if fields.StartAt != nil {
		start, err := a.when(*fields.StartAt, location)
		if err != nil {
			return err
		}
		body["start_at"] = start
		body["end_at"] = start.Add(keptDuration(stored, fields.IsAllDay))
	}
	if fields.EndAt != nil {
		end, err := a.when(*fields.EndAt, location)
		if err != nil {
			return err
		}
		body["end_at"] = end
	}
	return nil
}

func (a Agenda) when(value string, location *time.Location) (time.Time, error) {
	now := a.now().In(location)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	return parseWhen(value, today, location)
}
