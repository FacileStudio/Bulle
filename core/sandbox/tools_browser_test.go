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
	t.Cleanup(func() {
		if err := closer.Close(); err != nil {
			t.Errorf("closing: %v", err)
		}
	})
	return nacelle.ToolsByName(tools)
}

func TestBrowserScreenshotReturnsTheImage(t *testing.T) {
	var remote string
	tools := browserToolsWith(t, `{"ok":true,"text":"Screenshot of x","image":{"media_type":"image/jpeg","data":"anBn"}}`, &remote)
	result, err := tools["browser_screenshot"].Run(context.Background(), nacelle.Call{Input: json.RawMessage(`{}`)}, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	text, images := result.Text, result.Images
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
	_, err := tools["browser_confirm"].Run(context.Background(), nacelle.Call{Input: json.RawMessage(`{"held_id":"h9"}`)}, nil)
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
	if err := closer.Close(); err != nil {
		t.Fatalf("closing: %v", err)
	}
	for _, tool := range tools {
		if strings.HasPrefix(tool.Spec().Name, "browser_") {
			t.Fatalf("%s present without ToolsOptions.Browser", tool.Spec().Name)
		}
	}
}
