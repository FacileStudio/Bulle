package suite

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/nacelle"
)

type deadlineTool struct {
	spec   nacelle.Spec
	left   time.Duration
	text   string
	images []nacelle.Image
}

func (d *deadlineTool) Spec() nacelle.Spec { return d.spec }

func (d *deadlineTool) Run(ctx context.Context, _ nacelle.Call, _ func(string)) (nacelle.Result, error) {
	deadline, _ := ctx.Deadline()
	d.left = time.Until(deadline)
	return nacelle.Result{Text: d.text, Images: d.images}, nil
}

func TestAnMCPAnswerKeepsFewImagesUnderAMebibyte(t *testing.T) {
	image := func(size int) nacelle.Image { return nacelle.Image{MediaType: "image/png", Data: make([]byte, size)} }
	cases := []struct {
		images  []nacelle.Image
		kept    int
		dropped string
	}{
		{[]nacelle.Image{image(10), image(10)}, 2, ""},
		{[]nacelle.Image{image(10), image(10), image(10), image(10), image(10), image(10)}, 4, "[2 images left out"},
		{[]nacelle.Image{image(600 << 10), image(600 << 10), image(10)}, 1, "[2 images left out"},
	}
	for _, c := range cases {
		tool := &deadlineTool{spec: nacelle.Spec{Name: "toile_snapshot", ReadOnly: true}, images: c.images}
		result, _ := modelTools([]nacelle.Tool{tool})[0].Run(context.Background(), nacelle.Call{}, func(string) {})
		noted := strings.Contains(result.Text, "images left out")
		if len(result.Images) != c.kept || noted != (c.dropped != "") || !strings.Contains(result.Text, c.dropped) {
			t.Errorf("kept %d images, note %q; want %d and %q", len(result.Images), result.Text, c.kept, c.dropped)
		}
	}
}

func TestABoundCallIsTimedLoggedAndClipped(t *testing.T) {
	var logs bytes.Buffer
	m := NewMCP(Config{}, nil)
	m.SetLogger(slog.New(slog.NewTextHandler(&logs, nil)))
	read := &deadlineTool{spec: nacelle.Spec{Name: "docs_read", ReadOnly: true}, text: strings.Repeat("a", maxToolOutput+10)}
	write := &deadlineTool{spec: nacelle.Spec{Name: "docs_replace"}}
	tools := modelTools([]nacelle.Tool{m.bound(read, AppDocs, 7), m.bound(write, AppDocs, 7)})
	result, err := tools[0].Run(context.Background(), nacelle.Call{Name: "docs_read", Input: []byte(`{"secret":"hunter2"}`)}, func(string) {})
	if err != nil || len(result.Text) > maxToolOutput+100 || !strings.Contains(result.Text, "truncated") {
		t.Fatalf("an MCP answer must be clipped: %d bytes, %v", len(result.Text), err)
	}
	if _, err := tools[1].Run(context.Background(), nacelle.Call{Name: "docs_replace", Input: []byte(`{}`)}, func(string) {}); err != nil {
		t.Fatal(err)
	}
	if read.left > mcpReadLimit || read.left < mcpReadLimit-time.Second || write.left < mcpReadLimit {
		t.Fatalf("a read gets %v, a write that may be confirmed gets %v", read.left, write.left)
	}
	line := logs.String()
	if strings.Count(line, "msg=\"mcp call\"") != 2 || !strings.Contains(line, "tool=docs_read") || !strings.Contains(line, "user=7") ||
		!strings.Contains(line, "outcome=ok") || !strings.Contains(line, "args_sha256=") || strings.Contains(line, "hunter2") {
		t.Fatalf("one line per call, with a hash and never the arguments: %s", line)
	}
}

func TestAChangedOrUnknownToolIsLoggedOnce(t *testing.T) {
	var logs bytes.Buffer
	m := NewMCP(Config{}, nil)
	m.SetLogger(slog.New(slog.NewTextHandler(&logs, nil)))
	read := nacelle.Spec{Name: "docs_read", ReadOnly: true, Schema: map[string]any{"type": "object"}}
	m.watch(AppDocs, []nacelle.Spec{read, {Name: "docs_get"}})
	m.watch(AppDocs, []nacelle.Spec{read, {Name: "docs_get"}})
	if strings.Count(logs.String(), "unknown to Bulle") != 1 || strings.Contains(logs.String(), "changed") {
		t.Fatalf("an unknown tool is logged once and a stable one never: %s", logs.String())
	}
	read.ReadOnly = false
	m.watch(AppDocs, []nacelle.Spec{read})
	if !strings.Contains(logs.String(), "mcp tool definition changed") || !strings.Contains(logs.String(), "tool=docs_read") {
		t.Fatalf("a tool whose annotations changed must be logged: %s", logs.String())
	}
}
