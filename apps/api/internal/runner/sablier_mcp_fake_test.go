package runner_test

import (
	"context"
	"sync"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// fakeSablierMCP is Sablier's MCP server: the same nine tools, timer and
// entry writes that run at once, and an entry delete confirmed through MRTR
// the way tronc/mcp asks.
type fakeSablierMCP struct {
	*fakeSuiteMCP

	mu       sync.Mutex
	started  []int64
	deleted  []int64
	declined int
}

type timeIn struct {
	EntryID   int64  `json:"entry_id,omitempty"`
	ProjectID int64  `json:"project_id,omitempty"`
	TaskID    int64  `json:"task_id,omitempty"`
	Timezone  string `json:"timezone,omitempty"`
}

type timeOut struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Items   []any  `json:"items,omitempty"`
}

func newFakeSablierMCP(t *testing.T) *fakeSablierMCP {
	t.Helper()
	f := &fakeSablierMCP{}
	f.fakeSuiteMCP = newFakeSuiteMCP(t, suite.AppSablier, f.register)
	return f
}

func (f *fakeSablierMCP) register(server *sdk.Server) {
	read := func(item map[string]any) sdk.ToolHandlerFor[timeIn, timeOut] {
		return func(context.Context, *sdk.CallToolRequest, timeIn) (*sdk.CallToolResult, timeOut, error) {
			return nil, timeOut{Status: "ok", Items: []any{item}}, nil
		}
	}
	entry := map[string]any{"entry_id": 9, "project": "Web", "task": injected, "minutes": 30}
	sdk.AddTool(server, &sdk.Tool{Name: "sablier_timer_status", Annotations: hints(true, false, false)}, read(map[string]any{"running": false}))
	sdk.AddTool(server, &sdk.Tool{Name: "sablier_list_entries", Annotations: hints(true, false, true)}, read(entry))
	sdk.AddTool(server, &sdk.Tool{Name: "sablier_list_projects", Annotations: hints(true, false, true)}, read(map[string]any{"project_id": 2, "name": "Web"}))
	sdk.AddTool(server, &sdk.Tool{Name: "sablier_weekly_summary", Annotations: hints(true, false, true)}, read(entry))
	sdk.AddTool(server, &sdk.Tool{Name: "sablier_start_timer", Annotations: hints(false, false, false)}, f.start)
	for _, name := range []string{"sablier_stop_timer", "sablier_add_entry", "sablier_update_entry"} {
		sdk.AddTool(server, &sdk.Tool{Name: name, Annotations: hints(false, false, false)}, read(entry))
	}
	sdk.AddTool(server, &sdk.Tool{Name: "sablier_delete_entry", Annotations: hints(false, true, false)}, f.delete)
}

func (f *fakeSablierMCP) start(_ context.Context, _ *sdk.CallToolRequest, in timeIn) (*sdk.CallToolResult, timeOut, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.started = append(f.started, in.ProjectID)
	return nil, timeOut{Status: "started", Message: "The timer runs."}, nil
}

// delete asks first like Sablier: the first call returns input_required, and
// only an accept that keeps the box ticked deletes.
func (f *fakeSablierMCP) delete(_ context.Context, req *sdk.CallToolRequest, in timeIn) (*sdk.CallToolResult, timeOut, error) {
	if req.Params.RequestState == "" {
		question := &sdk.ElicitParams{Mode: "form", Message: "Delete the entry of 30 minutes on Web for good?", RequestedSchema: confirmForm}
		return &sdk.CallToolResult{InputRequests: sdk.InputRequestMap{"confirm": question}, RequestState: "1.entry-9.0.sig"}, timeOut{}, nil
	}
	answer, _ := req.Params.InputResponses["confirm"].(*sdk.ElicitResult)
	f.mu.Lock()
	defer f.mu.Unlock()
	if answer == nil || answer.Action != "accept" || answer.Content["confirm"] == false {
		f.declined++
		return nil, timeOut{Status: "declined", Message: "The user declined. The entry is kept."}, nil
	}
	f.deleted = append(f.deleted, in.EntryID)
	return nil, timeOut{Status: "deleted", Message: "The user confirmed and the entry was deleted."}, nil
}

func (f *fakeSablierMCP) counts() (started, deleted []int64, declined int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]int64(nil), f.started...), append([]int64(nil), f.deleted...), f.declined
}
