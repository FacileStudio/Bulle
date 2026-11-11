package suite

import (
	"context"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

func personalTools(t *testing.T, gw *Gateway, saved *store.Memory, defaults Defaults) []nacelle.Tool {
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
	gw, fake := myceliumWorld(t)
	saved := store.NewMemory()
	tools := personalTools(t, gw, saved, Defaults{})

	if _, err := runTool(t, tools, "memory_write", `{"path":"people/alice.md","content":"# Alice\n\nAlice likes tea."}`); err != nil {
		t.Fatal(err)
	}
	out, err := runTool(t, tools, "memory_read", `{"path":"people/alice.md"}`)
	if err != nil || !strings.HasPrefix(out, "---\ntitle: \"Alice\"\ntype: actor\n---\n\n# Alice") {
		t.Fatalf("a page without frontmatter must get one: %q %v", out, err)
	}
	again := personalTools(t, gw, saved, Defaults{})
	out, err = runTool(t, again, "memory_search", `{"query":"tea"}`)
	if err != nil || !strings.Contains(out, `"path":"people/alice.md"`) {
		t.Fatalf("search: %q %v", out, err)
	}
	spaces := fake.Spaces("1")
	id, _ := saved.MyceliumSpace(context.Background(), 1)
	if len(spaces) != 1 || spaces[id] != PersonalSpaceName {
		t.Fatalf("want one Bulle space, got %v (saved %q)", spaces, id)
	}
}

func TestMemoryToolsPreferTheBulleSpace(t *testing.T) {
	gw, fake := myceliumWorld(t)
	team := fake.AddSpace("1", "Team")
	tools := personalTools(t, gw, store.NewMemory(), Defaults{MyceliumSpace: team})
	if _, err := runTool(t, tools, "memory_write", `{"path":"work.md","content":"x"}`); err != nil {
		t.Fatal(err)
	}
	if _, ok := fake.Page(team, "work.md"); !ok || len(fake.Spaces("1")) != 1 {
		t.Fatalf("the space's own Mycelium space was not used: %v", fake.Spaces("1"))
	}
}

func TestMemoryRefusesPathsOutsideTheSpace(t *testing.T) {
	gw, fake := myceliumWorld(t)
	tools := personalTools(t, gw, store.NewMemory(), Defaults{})
	for _, path := range []string{"../secrets.md", "/etc/passwd.md", "a/../../b.md", `a\b.md`, ".hidden.md", "notes.txt", "C:/x.md", ""} {
		if _, err := runTool(t, tools, "memory_write", `{"path":"`+strings.ReplaceAll(path, `\`, `\\`)+`","content":"x"}`); err == nil {
			t.Errorf("%q was accepted", path)
		}
	}
	if _, err := runTool(t, tools, "memory_read", `{"path":"../../tokens.json"}`); err == nil {
		t.Error("read escaped the memory directory")
	}
	if calls := fake.Calls(); len(calls) != 0 {
		t.Fatalf("a refused path reached Mycelium: %+v", calls)
	}
}

func TestMemoryNeverFallsBackToTheCommonTree(t *testing.T) {
	gw, fake := myceliumWorld(t)
	fake.MakeAdmin("1")
	empty := Runtime{PersonalSpace: func(context.Context) (string, error) { return "", nil }}
	tools, _, _ := Tools(gw, 1, Defaults{}, empty)
	if _, err := runTool(t, tools, "memory_search", `{"query":"x"}`); err == nil {
		t.Fatal("an empty space id must not reach Mycelium")
	}
	if calls := fake.Calls(); len(calls) != 0 {
		t.Fatalf("Mycelium was called: %+v", calls)
	}
}

func TestMemoryWritesKeepEachUserToTheirOwnSpace(t *testing.T) {
	gw, fake := myceliumWorld(t)
	theirs := fake.AddSpace("2", "Bulle")
	m := NewMycelium(gw)
	if err := m.Write(context.Background(), 1, theirs, "x.md", "x"); err == nil {
		t.Fatal("user 1 wrote into user 2's space")
	}
	if _, ok := fake.Page(theirs, "x.md"); ok {
		t.Fatal("the page landed")
	}
}
