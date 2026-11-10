package runner_test

import (
	"context"
	"sync"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// fakeAgenda is Agenda's server: the same seven tools, creates and updates
// that run at once, and a delete confirmed through MRTR the way tronc/mcp
// asks, with a single ticked checkbox.
type fakeAgenda struct {
	*fakeSuiteMCP

	mu       sync.Mutex
	created  []string
	deleted  []int64
	declined int
	asked    []string
}

type eventIn struct {
	EventID     int64   `json:"event_id,omitempty"`
	CalendarID  int64   `json:"calendar_id,omitempty"`
	CalendarIDs []int64 `json:"calendar_ids,omitempty"`
	Title       string  `json:"title,omitempty"`
	Start       string  `json:"start,omitempty"`
	End         string  `json:"end,omitempty"`
	From        string  `json:"from,omitempty"`
	To          string  `json:"to,omitempty"`
	Timezone    string  `json:"timezone,omitempty"`
}

type eventOut struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Items   []any  `json:"items,omitempty"`
}

var confirmForm = map[string]any{
	"type":       "object",
	"properties": map[string]any{"confirm": map[string]any{"type": "boolean", "title": "Confirm", "default": true}},
	"required":   []string{"confirm"},
}

func newFakeAgenda(t *testing.T) *fakeAgenda {
	t.Helper()
	f := &fakeAgenda{}
	f.fakeSuiteMCP = newFakeSuiteMCP(t, suite.AppAgenda, f.register)
	return f
}

func (f *fakeAgenda) register(server *sdk.Server) {
	read := func(item map[string]any) sdk.ToolHandlerFor[eventIn, eventOut] {
		return func(context.Context, *sdk.CallToolRequest, eventIn) (*sdk.CallToolResult, eventOut, error) {
			return nil, eventOut{Status: "ok", Items: []any{item}}, nil
		}
	}
	calendar := map[string]any{"calendar_id": 3, "name": "Personal", "role": "owner", "personal": true}
	event := map[string]any{"event_id": 4, "calendar_id": 3, "title": injected, "start": "2026-11-12T14:00:00+01:00"}
	sdk.AddTool(server, &sdk.Tool{Name: "agenda_list_calendars", Annotations: hints(true, false, false)}, read(calendar))
	sdk.AddTool(server, &sdk.Tool{Name: "agenda_list_events", Annotations: hints(true, false, true)}, read(event))
	sdk.AddTool(server, &sdk.Tool{Name: "agenda_read_event", Annotations: hints(true, false, true)}, read(event))
	sdk.AddTool(server, &sdk.Tool{Name: "agenda_find_free_slots", Annotations: hints(true, false, false)}, read(map[string]any{"start": "2026-11-12T09:00:00+01:00", "minutes": 60}))
	sdk.AddTool(server, &sdk.Tool{Name: "agenda_create_event", Annotations: hints(false, false, false)}, f.create)
	sdk.AddTool(server, &sdk.Tool{Name: "agenda_update_event", Annotations: hints(false, false, false)}, read(event))
	sdk.AddTool(server, &sdk.Tool{Name: "agenda_delete_event", Annotations: hints(false, true, false)}, f.delete)
}

func (f *fakeAgenda) create(_ context.Context, _ *sdk.CallToolRequest, in eventIn) (*sdk.CallToolResult, eventOut, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.created = append(f.created, in.Title)
	return nil, eventOut{Status: "created", Message: "The event is in the calendar."}, nil
}

// delete asks first like Agenda: the first call returns input_required, the
// retry carries the answer, and only an accept that keeps the box ticked
// deletes.
func (f *fakeAgenda) delete(_ context.Context, req *sdk.CallToolRequest, in eventIn) (*sdk.CallToolResult, eventOut, error) {
	if req.Params.RequestState == "" {
		f.mu.Lock()
		f.asked = append(f.asked, req.ProtocolVersion())
		f.mu.Unlock()
		if caps := req.ClientCapabilities(); caps == nil || caps.Elicitation == nil {
			return nil, eventOut{Status: "pending_confirmation"}, nil
		}
		return askDelete(), eventOut{}, nil
	}
	answer, _ := req.Params.InputResponses["confirm"].(*sdk.ElicitResult)
	f.mu.Lock()
	defer f.mu.Unlock()
	if answer == nil || answer.Action != "accept" || answer.Content["confirm"] == false {
		f.declined++
		return nil, eventOut{Status: "declined", Message: "The user declined. Nothing happened, and the event is kept."}, nil
	}
	f.deleted = append(f.deleted, 4)
	return nil, eventOut{Status: "deleted", Message: "The user confirmed and the event was deleted."}, nil
}

func askDelete() *sdk.CallToolResult {
	question := &sdk.ElicitParams{Mode: "form", Message: `Delete the event "Standup" on Thu 12 Nov 14:00 for good?`, RequestedSchema: confirmForm}
	return &sdk.CallToolResult{InputRequests: sdk.InputRequestMap{"confirm": question}, RequestState: "1.event-4.0.sig"}
}

func (f *fakeAgenda) counts() (created []string, deleted []int64, declined int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.created...), append([]int64(nil), f.deleted...), f.declined
}
