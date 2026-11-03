package sandbox

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/FacileStudio/nacelle"
)

func browserToolsWith(t *testing.T, answer string, remote *string) map[string]nacelle.Tool {
	t.Helper()
	runner := &mockRunner{runFunc: func(_ context.Context, _ string, args ...string) ([]byte, error) {
		*remote = args[len(args)-1]
		return []byte("node warning\n" + answer + "\n"), nil
	}}
	tools, closer, err := RemoteTools(ToolsOptions{Target: &Target{Host: "127.0.0.1", Port: 2226, User: "boite"}, Runner: runner, Browser: true})
	if err != nil {
		t.Fatalf("RemoteTools: %v", err)
	}
	t.Cleanup(func() { _ = closer.Close() })
	return nacelle.ToolsByName(tools)
}

func TestBrowserScreenshotReturnsTheImage(t *testing.T) {
	var remote string
	tools := browserToolsWith(t, `{"ok":true,"text":"Screenshot of x","image":{"media_type":"image/jpeg","data":"anBn"}}`, &remote)
	text, images, err := tools["browser_screenshot"].(nacelle.ImageTool).RunImages(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("RunImages: %v", err)
	}
	if text != "Screenshot of x" || len(images) != 1 || string(images[0].Data) != "jpg" || images[0].MediaType != "image/jpeg" {
		t.Fatalf("text %q images %#v", text, images)
	}
	if !strings.Contains(remote, `node "$HOME"/.bulle/bin/bulle-browser.mjs screenshot`) {
		t.Fatalf("remote command = %q", remote)
	}
}

func TestBrowserFailureReachesTheModel(t *testing.T) {
	var remote string
	tools := browserToolsWith(t, `{"ok":false,"error":"no held request h9"}`, &remote)
	_, err := tools["browser_confirm"].Run(context.Background(), json.RawMessage(`{"held_id":"h9"}`))
	if err == nil || err.Error() != "no held request h9" {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(remote, `confirm '{"held_id":"h9"}'`) {
		t.Fatalf("remote command = %q", remote)
	}
}

func TestBrowserToolsNeedTheFlag(t *testing.T) {
	tools, closer, err := RemoteTools(ToolsOptions{Target: &Target{}, Runner: &mockRunner{}})
	if err != nil {
		t.Fatalf("RemoteTools: %v", err)
	}
	_ = closer.Close()
	for _, tool := range tools {
		if strings.HasPrefix(tool.Name(), "browser_") {
			t.Fatalf("%s present without ToolsOptions.Browser", tool.Name())
		}
	}
}
