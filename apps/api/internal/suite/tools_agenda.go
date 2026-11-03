package suite

import "context"

type agendaEventsInput struct {
	From     string `json:"from,omitempty" jsonschema:"description=Start: today, tomorrow, YYYY-MM-DD or RFC 3339 (default today)"`
	To       string `json:"to,omitempty" jsonschema:"description=End, same formats (default 7 days after from)"`
	Calendar int64  `json:"calendar,omitempty" jsonschema:"description=Calendar id; defaults to every calendar the user can see"`
	Timezone string `json:"timezone,omitempty" jsonschema:"description=IANA timezone for recurring events, e.g. Europe/Paris"`
}

type agendaCalendarsInput struct{}

type agendaEventInput struct {
	ID int64 `json:"id" jsonschema:"required,description=Event id"`
}

type agendaCreateInput struct {
	Calendar int64 `json:"calendar,omitempty" jsonschema:"description=Calendar id; defaults to the user's personal calendar"`
	AgendaEventFields
}

type agendaUpdateInput struct {
	ID       int64 `json:"id" jsonschema:"required,description=Event id"`
	Calendar int64 `json:"calendar,omitempty" jsonschema:"description=Move the event to this calendar id"`
	AgendaEventFields
}

func agendaTools(set *toolSet, a Agenda) {
	readTool(set, "agenda_events", "List the user's calendar events from Agenda.",
		func(ctx context.Context, in agendaEventsInput) (string, error) {
			query := AgendaQuery{Calendar: set.defaults.calendar(in.Calendar), From: in.From, To: in.To, Timezone: in.Timezone}
			return rawResult(a.Events(ctx, set.userID, query))
		})
	readTool(set, "agenda_calendars", "List the user's calendars in Agenda.",
		func(ctx context.Context, _ agendaCalendarsInput) (string, error) {
			return jsonResult(a.Calendars(ctx, set.userID))
		})
	readTool(set, "agenda_event", "Show one Agenda event with its attendees and recurrence.",
		func(ctx context.Context, in agendaEventInput) (string, error) {
			return rawResult(a.Event(ctx, set.userID, in.ID))
		})
	writeTool(set, "agenda_create_event", "Create an event in the user's Agenda calendar.",
		func(ctx context.Context, in agendaCreateInput) (string, error) {
			return rawResult(a.Create(ctx, set.userID, set.defaults.calendar(in.Calendar), in.AgendaEventFields))
		})
	writeTool(set, "agenda_update_event", "Change an Agenda event. Only the fields you pass change; moving the start keeps the duration.",
		func(ctx context.Context, in agendaUpdateInput) (string, error) {
			return rawResult(a.Update(ctx, set.userID, in.ID, in.Calendar, in.AgendaEventFields))
		})
	writeTool(set, "agenda_delete_event", "Delete an Agenda event, including every occurrence of a recurring one.",
		func(ctx context.Context, in agendaEventInput) (string, error) {
			return done("deleted", a.Delete(ctx, set.userID, in.ID))
		})
}
