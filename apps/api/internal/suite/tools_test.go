package suite

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/core/nacelletest"
	"github.com/FacileStudio/nacelle"
)

func runTool(t *testing.T, tools []nacelle.Tool, name, input string) (string, error) {
	t.Helper()
	for _, tool := range tools {
		if tool.Spec().Name == name {
			return nacelletest.Run(context.Background(), tool, json.RawMessage(input))
		}
	}
	t.Fatalf("tool %s is not registered", name)
	return "", nil
}

func TestToolsRegisterOnlyConfiguredApps(t *testing.T) {
	w := newWorld(t)
	tools, note, err := Tools(everyApp(w), 1, Defaults{}, Runtime{})
	if err != nil || len(tools) == 0 {
		t.Fatalf("want Mycelium's tools, got %d, %v", len(tools), err)
	}
	for _, tool := range tools {
		if name := tool.Spec().Name; !strings.HasPrefix(name, "memory_") && !strings.HasPrefix(name, "artifact_") {
			t.Errorf("%s is registered, but only Mycelium has built-in tools", name)
		}
	}
	for _, name := range []string{"memory_search", "artifact_create"} {
		if !strings.Contains(note, name) {
			t.Errorf("prompt must list %s: %q", name, note)
		}
	}
	if strings.Contains(note, "glouton_") || strings.Contains(note, "sablier_") {
		t.Errorf("prompt lists a tool that is not registered: %q", note)
	}
}

func TestEveryMutatingBuiltInToolIsASelfWrite(t *testing.T) {
	w := newWorld(t)
	tools, _, err := Tools(everyApp(w), 1, Defaults{}, Runtime{})
	if err != nil {
		t.Fatal(err)
	}
	self := map[string]bool{}
	for _, name := range SelfWrites() {
		self[name] = true
	}
	for _, tool := range tools {
		mutates := !tool.Spec().ReadOnly
		if mutates != self[tool.Spec().Name] {
			t.Errorf("%s: mutates=%t but self write=%t", tool.Spec().Name, mutates, self[tool.Spec().Name])
		}
		delete(self, tool.Spec().Name)
	}
	for name := range self {
		t.Errorf("%s is a self write but never registered", name)
	}
}

func TestToolsNoneConfigured(t *testing.T) {
	w := newWorld(t)
	for _, name := range appNames {
		w.cfg.Apps[name] = AppConfig{}
	}
	tools, note, err := Tools(NewGateway(w.tokens(), w.cfg, nil), 1, Defaults{}, Runtime{})
	if err != nil || len(tools) != 0 || note != NoMemoryNote {
		t.Fatalf("want no tools and only the no-memory note, got %d tools, %q, %v", len(tools), note, err)
	}
}

func TestToolReportsReconnect(t *testing.T) {
	w := newWorld(t)
	tools, _, _ := Tools(everyApp(w), 99, Defaults{MyceliumSpace: "sp-team"}, Runtime{})
	_, err := runTool(t, tools, "memory_search", `{"query":"plans"}`)
	if err == nil || !strings.Contains(err.Error(), "sign in again") {
		t.Fatalf("want a reconnect message, got %v", err)
	}
}

func TestClipTruncates(t *testing.T) {
	long := strings.Repeat("a", maxToolOutput+10)
	if out := clip(long); !strings.Contains(out, "truncated") || len(out) > maxToolOutput+100 {
		t.Fatalf("clip did not truncate: %d", len(out))
	}
}
