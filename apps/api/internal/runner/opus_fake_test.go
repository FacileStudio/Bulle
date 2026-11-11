package runner_test

import (
	"context"
	"sync"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// fakeOpus is Opus's server: the same eleven tools, writes that run at once,
// and deletes confirmed through MRTR the way tronc/mcp asks.
type fakeOpus struct {
	*fakeSuiteMCP

	mu       sync.Mutex
	created  []string
	deleted  []int64
	declined int
}

type taskIn struct {
	TaskID    int64  `json:"task_id,omitempty"`
	ProjectID int64  `json:"project_id,omitempty"`
	Title     string `json:"title,omitempty"`
	Assignee  string `json:"assignee,omitempty"`
	Text      string `json:"text,omitempty"`
}

type taskOut struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Tasks   []any  `json:"tasks,omitempty"`
}

func newFakeOpus(t *testing.T) *fakeOpus {
	t.Helper()
	f := &fakeOpus{}
	f.fakeSuiteMCP = newFakeSuiteMCP(t, suite.AppOpus, f.register)
	return f
}

func (f *fakeOpus) register(server *sdk.Server) {
	read := func(item map[string]any) sdk.ToolHandlerFor[taskIn, taskOut] {
		return func(context.Context, *sdk.CallToolRequest, taskIn) (*sdk.CallToolResult, taskOut, error) {
			return nil, taskOut{Status: "ok", Tasks: []any{item}}, nil
		}
	}
	task := map[string]any{"task_id": 42, "ref": "web-12", "title": injected, "status": "to-do"}
	sdk.AddTool(server, &sdk.Tool{Name: "opus_list_workspaces", Annotations: hints(true, false, false)}, read(map[string]any{"workspace_id": 1, "name": "Personal"}))
	sdk.AddTool(server, &sdk.Tool{Name: "opus_list_projects", Annotations: hints(true, false, true)}, read(map[string]any{"project_id": 5, "name": "Web"}))
	sdk.AddTool(server, &sdk.Tool{Name: "opus_search_tasks", Annotations: hints(true, false, true)}, read(task))
	sdk.AddTool(server, &sdk.Tool{Name: "opus_read_task", Annotations: hints(true, false, true)}, read(task))
	sdk.AddTool(server, &sdk.Tool{Name: "opus_create_task", Annotations: hints(false, false, false)}, f.create)
	for _, name := range []string{"opus_update_task", "opus_comment_task", "opus_create_project", "opus_update_project"} {
		sdk.AddTool(server, &sdk.Tool{Name: name, Annotations: hints(false, false, false)}, read(task))
	}
	sdk.AddTool(server, &sdk.Tool{Name: "opus_delete_task", Annotations: hints(false, true, false)}, f.delete)
	sdk.AddTool(server, &sdk.Tool{Name: "opus_delete_project", Annotations: hints(false, true, false)}, f.delete)
}

func (f *fakeOpus) create(_ context.Context, _ *sdk.CallToolRequest, in taskIn) (*sdk.CallToolResult, taskOut, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.created = append(f.created, in.Title)
	return nil, taskOut{Status: "created", Message: "The task is in the project."}, nil
}

// delete asks first like Opus: the first call returns input_required, and
// only an accept that keeps the box ticked deletes.
func (f *fakeOpus) delete(_ context.Context, req *sdk.CallToolRequest, in taskIn) (*sdk.CallToolResult, taskOut, error) {
	if req.Params.RequestState == "" {
		question := &sdk.ElicitParams{Mode: "form", Message: `Delete the task "web-12 Fix the export" for good?`, RequestedSchema: confirmForm}
		return &sdk.CallToolResult{InputRequests: sdk.InputRequestMap{"confirm": question}, RequestState: "1.task-42.0.sig"}, taskOut{}, nil
	}
	answer, _ := req.Params.InputResponses["confirm"].(*sdk.ElicitResult)
	f.mu.Lock()
	defer f.mu.Unlock()
	if answer == nil || answer.Action != "accept" || answer.Content["confirm"] == false {
		f.declined++
		return nil, taskOut{Status: "declined", Message: "The user declined. Nothing was deleted."}, nil
	}
	f.deleted = append(f.deleted, in.TaskID)
	return nil, taskOut{Status: "deleted", Message: "The user confirmed and the task was deleted."}, nil
}

func (f *fakeOpus) counts() (created []string, deleted []int64, declined int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.created...), append([]int64(nil), f.deleted...), f.declined
}
