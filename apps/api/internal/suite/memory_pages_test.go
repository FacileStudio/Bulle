package suite

import (
	"context"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

func TestListPagesPutsProfileFirstAndHidesTheRest(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	w.fakes.own("sp", "1")
	for _, path := range []string{"zeta.md", "profile.md", "people/ann.md"} {
		w.fakes.pages[fakePage{space: "sp", path: path}] = "x"
	}
	pages, err := NewMycelium(NewGateway(w.tokens(), w.cfg, nil)).ListPages(context.Background(), 1, "sp")
	if err != nil || len(pages) != 3 {
		t.Fatalf("pages %+v, %v", pages, err)
	}
	if pages[0].Path != "profile.md" || pages[1].Path != "people/ann.md" || pages[2].Path != "zeta.md" {
		t.Fatalf("order %+v", pages)
	}
	if _, err := NewMycelium(NewGateway(w.tokens(), w.cfg, nil)).ListPages(context.Background(), 2, "sp"); err == nil {
		t.Fatal("another user listed the space")
	}
}

func TestMemoryDeleteRemovesOnePage(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	tools := personalTools(t, w, NewGateway(w.tokens(), w.cfg, nil), store.NewMemory(), Defaults{})
	if _, err := runTool(t, tools, "memory_write", `{"path":"people/bob.md","content":"Bob."}`); err != nil {
		t.Fatal(err)
	}
	if out, err := runTool(t, tools, "memory_delete", `{"path":"people/bob.md"}`); err != nil || out != "Deleted people/bob.md from memory." {
		t.Fatalf("delete: %q %v", out, err)
	}
	if _, err := runTool(t, tools, "memory_read", `{"path":"people/bob.md"}`); err == nil {
		t.Fatal("the page is still there")
	}
	for _, path := range []string{"../x.md", ".hidden.md", "notes.txt", ""} {
		if _, err := runTool(t, tools, "memory_delete", `{"path":"`+path+`"}`); err == nil {
			t.Errorf("%q was deleted", path)
		}
	}
}
