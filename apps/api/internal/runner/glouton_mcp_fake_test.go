package runner_test

import (
	"context"
	"sync"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// fakeGloutonMCP is Glouton's MCP server: the same fourteen tools, writes
// that run at once, and the two deletes confirmed through MRTR the way
// tronc's kit asks.
type fakeGloutonMCP struct {
	*fakeSuiteMCP

	mu       sync.Mutex
	created  []string
	deleted  []string
	declined int
}

type contactIn struct {
	Query     string `json:"query,omitempty"`
	Name      string `json:"name,omitempty"`
	CompanyID string `json:"company_id,omitempty"`
	PersonID  string `json:"person_id,omitempty"`
}

type contactOut struct {
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	Items   []any  `json:"items,omitempty"`
	Company any    `json:"company,omitempty"`
}

func newFakeGloutonMCP(t *testing.T) *fakeGloutonMCP {
	t.Helper()
	f := &fakeGloutonMCP{}
	f.fakeSuiteMCP = newFakeSuiteMCP(t, suite.AppGlouton, f.register)
	return f
}

func (f *fakeGloutonMCP) register(server *sdk.Server) {
	read := func(item map[string]any) sdk.ToolHandlerFor[contactIn, contactOut] {
		return func(context.Context, *sdk.CallToolRequest, contactIn) (*sdk.CallToolResult, contactOut, error) {
			return nil, contactOut{Items: []any{item}}, nil
		}
	}
	company := map[string]any{"company_id": "c1", "name": "Acme", "notes": injected}
	for _, name := range []string{"glouton_contacts_search", "glouton_company_get", "glouton_person_get", "glouton_outreach_list", "glouton_suggestions_list"} {
		sdk.AddTool(server, &sdk.Tool{Name: name, Annotations: hints(true, false, false)}, read(company))
	}
	sdk.AddTool(server, &sdk.Tool{Name: "glouton_company_create", Annotations: hints(false, false, false)}, f.create)
	for _, name := range []string{"glouton_company_update", "glouton_person_create", "glouton_person_update", "glouton_draft_save", "glouton_draft_update", "glouton_suggestion_create"} {
		sdk.AddTool(server, &sdk.Tool{Name: name, Annotations: hints(false, false, false)}, read(company))
	}
	sdk.AddTool(server, &sdk.Tool{Name: "glouton_company_delete", Annotations: hints(false, true, false)}, f.delete)
	sdk.AddTool(server, &sdk.Tool{Name: "glouton_person_delete", Annotations: hints(false, true, false)}, f.delete)
}

func (f *fakeGloutonMCP) create(_ context.Context, _ *sdk.CallToolRequest, in contactIn) (*sdk.CallToolResult, contactOut, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.created = append(f.created, in.Name)
	return nil, contactOut{Company: map[string]any{"id": "c2", "name": in.Name}}, nil
}

// delete asks first like Glouton: the first call returns input_required, and
// only an accept that keeps the box ticked deletes.
func (f *fakeGloutonMCP) delete(_ context.Context, req *sdk.CallToolRequest, in contactIn) (*sdk.CallToolResult, contactOut, error) {
	if req.Params.RequestState == "" {
		question := &sdk.ElicitParams{Mode: "form", Message: `Delete the company "Acme" and its people from Glouton?`, RequestedSchema: confirmForm}
		return &sdk.CallToolResult{InputRequests: sdk.InputRequestMap{"confirm": question}, RequestState: "1.company-c1.0.sig"}, contactOut{}, nil
	}
	answer, _ := req.Params.InputResponses["confirm"].(*sdk.ElicitResult)
	f.mu.Lock()
	defer f.mu.Unlock()
	if answer == nil || answer.Action != "accept" || answer.Content["confirm"] == false {
		f.declined++
		return nil, contactOut{Status: "declined", Message: "The user declined. Nothing was deleted."}, nil
	}
	f.deleted = append(f.deleted, in.CompanyID)
	return nil, contactOut{Status: "done", Message: "The company was deleted."}, nil
}

func (f *fakeGloutonMCP) counts() (created, deleted []string, declined int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.created...), append([]string(nil), f.deleted...), f.declined
}

func (f *fakeGloutonMCP) throttleCalls(n int) {
	f.fakeSuiteMCP.mu.Lock()
	defer f.fakeSuiteMCP.mu.Unlock()
	f.throttle = n
}
