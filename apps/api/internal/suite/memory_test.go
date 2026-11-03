package suite

import (
	"context"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

func personalTools(t *testing.T, w *world, gw *Gateway, saved *store.Memory, defaults Defaults) []nacelle.Tool {
	t.Helper()
	rt := Runtime{PersonalSpace: func(ctx context.Context) (string, error) {
		return NewMycelium(gw).PersonalSpace(ctx, 1, saved)
	}}
	tools, note, err := Tools(gw, 1, defaults, rt)
	if err != nil || !strings.Contains(note, "long-term memory") {
		t.Fatalf("tools: %v, note %q", err, note)
	}
	return tools
}

func TestMemoryToolsCreateAndReuseThePersonalSpace(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	gw := NewGateway(w.tokens(), w.cfg, nil)
	saved := store.NewMemory()
	tools := personalTools(t, w, gw, saved, Defaults{})

	if _, err := runTool(t, tools, "memory_write", `{"path":"people/alice.md","content":"Alice likes tea."}`); err != nil {
		t.Fatal(err)
	}
	if out, err := runTool(t, tools, "memory_read", `{"path":"people/alice.md"}`); err != nil || out != "Alice likes tea." {
		t.Fatalf("read: %q %v", out, err)
	}
	again := personalTools(t, w, gw, saved, Defaults{})
	out, err := runTool(t, again, "memory_search", `{"query":"tea"}`)
	if err != nil || !strings.Contains(out, `"path":"people/alice.md"`) {
		t.Fatalf("search: %q %v", out, err)
	}
	if id, _ := saved.MyceliumSpace(context.Background(), 1); id != "sp-1-1" || w.fakes.creates != 1 {
		t.Fatalf("space %q created %d times", id, w.fakes.creates)
	}
}

func TestMemoryToolsPreferTheBulleSpace(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	gw := NewGateway(w.tokens(), w.cfg, nil)
	w.fakes.own("sp-team", "1")
	tools := personalTools(t, w, gw, store.NewMemory(), Defaults{MyceliumSpace: "sp-team"})
	if _, err := runTool(t, tools, "memory_write", `{"path":"work.md","content":"x"}`); err != nil {
		t.Fatal(err)
	}
	if _, ok := w.fakes.pages[fakePage{space: "sp-team", path: "work.md"}]; !ok || w.fakes.creates != 0 {
		t.Fatalf("the space's own Mycelium space was not used: %+v", w.fakes.pages)
	}
}

func TestMemoryRefusesPathsOutsideTheSpace(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	tools := personalTools(t, w, NewGateway(w.tokens(), w.cfg, nil), store.NewMemory(), Defaults{})
	for _, path := range []string{"../secrets.md", "/etc/passwd.md", "a/../../b.md", `a\b.md`, ".hidden.md", "notes.txt", "C:/x.md", ""} {
		if _, err := runTool(t, tools, "memory_write", `{"path":"`+strings.ReplaceAll(path, `\`, `\\`)+`","content":"x"}`); err == nil {
			t.Errorf("%q was accepted", path)
		}
	}
	if _, err := runTool(t, tools, "memory_read", `{"path":"../../tokens.json"}`); err == nil {
		t.Error("read escaped the memory directory")
	}
	if len(w.writes) != 0 {
		t.Fatalf("a refused path reached Mycelium: %+v", w.writes)
	}
}

func TestMemoryNeverFallsBackToTheCommonTree(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	empty := Runtime{PersonalSpace: func(context.Context) (string, error) { return "", nil }}
	tools, _, _ := Tools(NewGateway(w.tokens(), w.cfg, nil), 1, Defaults{}, empty)
	if _, err := runTool(t, tools, "memory_search", `{"query":"x"}`); err == nil {
		t.Fatal("an empty space id must not reach Mycelium")
	}
	if w.count(&w.appCalls) != 0 {
		t.Fatalf("Mycelium was called %d times", w.appCalls)
	}
}
