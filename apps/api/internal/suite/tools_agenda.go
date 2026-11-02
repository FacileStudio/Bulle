package suite

import "context"

type agendaEventsInput struct {
	From     string `json:"from,omitempty" jsonschema:"description=Start: today, tomorrow, YYYY-MM-DD or RFC 3339 (default today)"`
	To       string `json:"to,omitempty" jsonschema:"description=End, same formats (default 7 days after from)"`
	Calendar int64  `json:"calendar,omitempty" jsonschema:"description=Calendar id; defaults to every calendar the user can see"`
	Timezone string `json:"timezone,omitempty" jsonschema:"description=IANA timezone for recurring events, e.g. Europe/Paris"`
}

type agendaCalendarsInput struct{}

func agendaTools(set *toolSet, a Agenda) {
	readTool(set, "agenda_events", "List the user's calendar events from Agenda.",
		func(ctx context.Context, in agendaEventsInput) (string, error) {
			query := AgendaQuery{Calendar: in.Calendar, From: in.From, To: in.To, Timezone: in.Timezone}
			return rawResult(a.Events(ctx, set.userID, query))
		})
	readTool(set, "agenda_calendars", "List the user's calendars in Agenda.",
		func(ctx context.Context, _ agendaCalendarsInput) (string, error) {
			return jsonResult(a.Calendars(ctx, set.userID))
		})
}
