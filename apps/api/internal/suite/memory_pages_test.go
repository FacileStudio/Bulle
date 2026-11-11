package suite

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle/mcp/client"
)

func TestListPagesPutsProfileFirst(t *testing.T) {
	gw, fake := myceliumWorld(t)
	space := fake.AddSpace("1", "Bulle")
	for _, path := range []string{"zeta.md", "profile.md", "people/ann.md"} {
		fake.SetPage(space, path, "x")
	}
	pages, err := NewMycelium(gw).ListPages(context.Background(), 1, space)
	if err != nil || len(pages) != 3 {
		t.Fatalf("pages %+v, %v", pages, err)
	}
	if pages[0].Path != "profile.md" || pages[1].Path != "people/ann.md" || pages[2].Path != "zeta.md" {
		t.Fatalf("order %+v", pages)
	}
	if _, err := NewMycelium(gw).ListPages(context.Background(), 2, space); err == nil {
		t.Fatal("another user listed the space")
	}
}

func TestMemoryDeleteWaitsOnMyceliumsConfirmation(t *testing.T) {
	gw, fake := myceliumWorld(t)
	saved := store.NewMemory()
	tools := personalTools(t, gw, saved, Defaults{})
	if _, err := runTool(t, tools, "memory_write", `{"path":"people/bob.md","content":"Bob."}`); err != nil {
		t.Fatal(err)
	}
	space, _ := saved.MyceliumSpace(context.Background(), 1)
	if _, err := runTool(t, tools, "memory_delete", `{"path":"people/bob.md"}`); !errors.Is(err, ErrDeclined) {
		t.Fatalf("without anyone to answer, Mycelium's question must decline: %v", err)
	}
	if _, ok := fake.Page(space, "people/bob.md"); !ok {
		t.Fatal("a delete nobody confirmed removed the page")
	}
	asked := ""
	ctx := WithElicit(context.Background(), func(_ context.Context, request client.Elicitation) (client.Answer, error) {
		asked = request.Request.Message
		content, err := Acceptance(request.Request.RequestedSchema)
		return client.Answer{Accept: true, Content: content}, err
	})
	if err := NewMycelium(gw).Delete(ctx, 1, space, "people/bob.md"); err != nil || !strings.Contains(asked, "people/bob.md") {
		t.Fatalf("an accepted delete must run once: %v, asked %q", err, asked)
	}
	if _, ok := fake.Page(space, "people/bob.md"); ok {
		t.Fatal("the page is still there")
	}
	for _, path := range []string{"../x.md", ".hidden.md", "notes.txt", ""} {
		if _, err := runTool(t, tools, "memory_delete", `{"path":"`+path+`"}`); err == nil {
			t.Errorf("%q was deleted", path)
		}
	}
}
