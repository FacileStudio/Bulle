package suite

import "context"

type agendaEventsInput struct {
	From     string `json:"from,omitempty" jsonschema:"description=Start: today, tomorrow, YYYY-MM-DD or RFC 3339 (default today)"`
	To       string `json:"to,omitempty" jsonschema:"description=End, same formats (default 7 days after from)"`
	Calendar int64  `json:"calendar,omitempty" jsonschema:"description=Calendar id from agenda_list_calendars; defaults to every calendar the user can see"`
	Timezone string `json:"timezone,omitempty" jsonschema:"description=IANA timezone for recurring events, e.g. Europe/Paris"`
}

type agendaCalendarsInput struct{}

type agendaEventInput struct {
	EventID int64 `json:"event_id" jsonschema:"required,description=Event id as agenda_list_events answers it"`
}

type agendaCreateInput struct {
	Calendar int64 `json:"calendar,omitempty" jsonschema:"description=Calendar id from agenda_list_calendars; defaults to the user's personal calendar"`
	AgendaEventFields
}

type agendaUpdateInput struct {
	EventID  int64 `json:"event_id" jsonschema:"required,description=Event id as agenda_list_events answers it"`
	Calendar int64 `json:"calendar,omitempty" jsonschema:"description=Move the event to this calendar id"`
	AgendaEventFields
}

func agendaTools(set *toolSet, a Agenda) {
	readTool(set, "agenda_list_events", "List the user's calendar events in Agenda between from and to, by default the next 7 days. "+
		"Use it to answer what is coming up or to find an event's id before changing it. "+
		"Pass timezone, such as Europe/Paris, so day boundaries and recurring events follow the user's zone.",
		func(ctx context.Context, in agendaEventsInput) (string, error) {
			query := AgendaQuery{Calendar: set.defaults.calendar(in.Calendar), From: in.From, To: in.To, Timezone: in.Timezone}
			return rawResult(a.Events(ctx, set.userID, query))
		})
	readTool(set, "agenda_list_calendars", "List the user's Agenda calendars with their id, name, role and which one is personal. "+
		"Use it to find the calendar id other Agenda tools take.",
		func(ctx context.Context, _ agendaCalendarsInput) (string, error) {
			return jsonResult(a.Calendars(ctx, set.userID))
		})
	readTool(set, "agenda_get_event", "Show one Agenda event with its attendees and recurrence rule. "+
		"Use it before changing an event, to read its current state. event_id comes from agenda_list_events.",
		func(ctx context.Context, in agendaEventInput) (string, error) {
			return rawResult(a.Event(ctx, set.userID, in.EventID))
		})
	writeTool(set, "agenda_create_event", "Create an event in the user's Agenda, in calendar or else the personal calendar. "+
		"Pass start_at as RFC 3339 with an offset, or YYYY-MM-DD for an all-day event; end_at defaults to one hour later. "+
		"timezone is the IANA zone the event is anchored in, such as Europe/Paris.",
		func(ctx context.Context, in agendaCreateInput) (string, error) {
			return rawResult(a.Create(ctx, set.userID, set.defaults.calendar(in.Calendar), in.AgendaEventFields))
		})
	writeTool(set, "agenda_update_event", "Change an Agenda event. Only the fields you pass change, and moving the start keeps the duration. "+
		"attendees replaces the whole list. event_id comes from agenda_list_events.",
		func(ctx context.Context, in agendaUpdateInput) (string, error) {
			return rawResult(a.Update(ctx, set.userID, in.EventID, in.Calendar, in.AgendaEventFields))
		})
	writeTool(set, "agenda_delete_event", "Delete an Agenda event, including every occurrence of a recurring one. "+
		"Use it only when the user asks to delete or cancel the event. event_id comes from agenda_list_events.",
		func(ctx context.Context, in agendaEventInput) (string, error) {
			return done("deleted", a.Delete(ctx, set.userID, in.EventID))
		})
}
