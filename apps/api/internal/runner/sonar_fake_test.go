package runner_test

import (
	"context"
	"sync"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// fakeSonar is sonar's server: sonar_search and sonar_fetch, both read-only
// and open world, and a daily quota that answers isError once spent.
type fakeSonar struct {
	*fakeSuiteMCP

	mu       sync.Mutex
	left     int
	searched []string
	fetched  []string
}

type sonarIn struct {
	Query  string `json:"query,omitempty"`
	Kind   string `json:"kind,omitempty"`
	Count  int    `json:"count,omitempty"`
	URL    string `json:"url,omitempty"`
	Offset int    `json:"offset,omitempty"`
}

type sonarOut struct {
	Query      string `json:"query,omitempty"`
	Results    []any  `json:"results,omitempty"`
	URL        string `json:"url,omitempty"`
	Markdown   string `json:"markdown,omitempty"`
	NextOffset int    `json:"next_offset,omitempty"`
}

func newFakeSonar(t *testing.T, quota int) *fakeSonar {
	t.Helper()
	f := &fakeSonar{left: quota}
	f.fakeSuiteMCP = newFakeSuiteMCP(t, suite.AppSonar, f.register)
	return f
}

func (f *fakeSonar) register(server *sdk.Server) {
	sdk.AddTool(server, &sdk.Tool{Name: "sonar_search", Annotations: hints(true, false, true)}, f.search)
	sdk.AddTool(server, &sdk.Tool{Name: "sonar_fetch", Annotations: hints(true, false, true)}, f.fetch)
}

func (f *fakeSonar) search(_ context.Context, _ *sdk.CallToolRequest, in sonarIn) (*sdk.CallToolResult, sonarOut, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.left <= 0 {
		return &sdk.CallToolResult{IsError: true, Content: []sdk.Content{&sdk.TextContent{
			Text: "You have used today's 20 free searches. Upgrade at https://sso.facile.studio/spaces"}}}, sonarOut{}, nil
	}
	f.left--
	f.searched = append(f.searched, in.Query)
	return nil, sonarOut{Query: in.Query, Results: []any{map[string]any{"title": "Go 1.27", "url": "https://go.dev/doc/go1.27", "snippet": injected}}}, nil
}

func (f *fakeSonar) fetch(_ context.Context, _ *sdk.CallToolRequest, in sonarIn) (*sdk.CallToolResult, sonarOut, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fetched = append(f.fetched, in.URL)
	return nil, sonarOut{URL: in.URL, Markdown: "# Go 1.27\n\n" + injected, NextOffset: 4000}, nil
}

func (f *fakeSonar) counts() (searched, fetched []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.searched...), append([]string(nil), f.fetched...)
}
