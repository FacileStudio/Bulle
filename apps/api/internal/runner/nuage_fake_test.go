package runner_test

import (
	"context"
	"sync"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// fakeNuage is Nuage's server: the same ten tools, saves that run at once,
// and a trash or a public link confirmed through MRTR the way tronc/mcp asks.
type fakeNuage struct {
	*fakeSuiteMCP

	mu       sync.Mutex
	saved    []string
	trashed  []int64
	declined int
}

type fileIn struct {
	Type     string `json:"type,omitempty"`
	ID       int64  `json:"id,omitempty"`
	FileID   int64  `json:"file_id,omitempty"`
	FolderID int64  `json:"folder_id,omitempty"`
	Name     string `json:"name,omitempty"`
	Content  string `json:"content,omitempty"`
	Query    string `json:"query,omitempty"`
}

type fileOut struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Items   []any  `json:"items,omitempty"`
}

func newFakeNuage(t *testing.T) *fakeNuage {
	t.Helper()
	f := &fakeNuage{}
	f.fakeSuiteMCP = newFakeSuiteMCP(t, suite.AppNuage, f.register)
	return f
}

func (f *fakeNuage) register(server *sdk.Server) {
	read := func(item map[string]any) sdk.ToolHandlerFor[fileIn, fileOut] {
		return func(context.Context, *sdk.CallToolRequest, fileIn) (*sdk.CallToolResult, fileOut, error) {
			return nil, fileOut{Status: "ok", Items: []any{item}}, nil
		}
	}
	file := map[string]any{"type": "file", "id": 12, "name": "notes.md", "content": injected}
	sdk.AddTool(server, &sdk.Tool{Name: "nuage_list_folder", Annotations: hints(true, false, true)}, read(file))
	sdk.AddTool(server, &sdk.Tool{Name: "nuage_search", Annotations: hints(true, false, true)}, read(file))
	sdk.AddTool(server, &sdk.Tool{Name: "nuage_read_file", Annotations: hints(true, false, true)}, read(file))
	sdk.AddTool(server, &sdk.Tool{Name: "nuage_list_shares", Annotations: hints(true, false, false)}, read(map[string]any{"share_id": 2}))
	sdk.AddTool(server, &sdk.Tool{Name: "nuage_save_file", Annotations: hints(false, true, false)}, f.save)
	sdk.AddTool(server, &sdk.Tool{Name: "nuage_create_folder", Annotations: hints(false, false, false)}, read(map[string]any{"type": "folder", "id": 3}))
	sdk.AddTool(server, &sdk.Tool{Name: "nuage_move", Annotations: hints(false, false, false)}, read(file))
	sdk.AddTool(server, &sdk.Tool{Name: "nuage_revoke_share", Annotations: hints(false, true, false)}, read(map[string]any{"share_id": 2}))
	sdk.AddTool(server, &sdk.Tool{Name: "nuage_delete", Annotations: hints(false, true, false)}, f.trash)
	sdk.AddTool(server, &sdk.Tool{Name: "nuage_create_share", Annotations: hints(false, false, true)}, f.trash)
}

func (f *fakeNuage) save(_ context.Context, _ *sdk.CallToolRequest, in fileIn) (*sdk.CallToolResult, fileOut, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.saved = append(f.saved, in.Name)
	return nil, fileOut{Status: "created", Message: "The file is saved."}, nil
}

// trash asks first like Nuage: the first call returns input_required, and
// only an accept that keeps the box ticked moves the item.
func (f *fakeNuage) trash(_ context.Context, req *sdk.CallToolRequest, in fileIn) (*sdk.CallToolResult, fileOut, error) {
	if req.Params.RequestState == "" {
		question := &sdk.ElicitParams{Mode: "form", Message: `Move the file "notes.md" to the trash?`, RequestedSchema: confirmForm}
		return &sdk.CallToolResult{InputRequests: sdk.InputRequestMap{"confirm": question}, RequestState: "1.file-12.0.sig"}, fileOut{}, nil
	}
	answer, _ := req.Params.InputResponses["confirm"].(*sdk.ElicitResult)
	f.mu.Lock()
	defer f.mu.Unlock()
	if answer == nil || answer.Action != "accept" || answer.Content["confirm"] == false {
		f.declined++
		return nil, fileOut{Status: "declined", Message: "The user declined. Nothing changed."}, nil
	}
	f.trashed = append(f.trashed, in.ID)
	return nil, fileOut{Status: "trashed", Message: "The user confirmed and the file is in the trash."}, nil
}

func (f *fakeNuage) counts() (saved []string, trashed []int64, declined int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.saved...), append([]int64(nil), f.trashed...), f.declined
}
