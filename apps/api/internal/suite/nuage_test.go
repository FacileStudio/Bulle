package suite

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func nuageWorld(t *testing.T) (*world, Runtime) {
	t.Helper()
	w := newWorld(t)
	w.addUser(1, future)
	rt := Runtime{Attachment: func(_ context.Context, id string) (File, error) {
		if id != "att-1" {
			return File{}, errors.New("not found")
		}
		return File{Name: "scan.pdf", MediaType: "application/pdf", Data: []byte("%PDF-1.7")}, nil
	}}
	return w, rt
}

func TestFilesReadTools(t *testing.T) {
	w, rt := nuageWorld(t)
	tools, _, err := Tools(NewGateway(w.tokens(), w.cfg, nil), 1, Defaults{}, rt)
	if err != nil {
		t.Fatal(err)
	}
	listing, _ := runTool(t, tools, "nuage_list", `{}`)
	if !strings.Contains(listing, `"name":"Docs"`) || !strings.Contains(listing, `"name":"notes.txt"`) {
		t.Errorf("nuage_list: %s", listing)
	}
	if hits, _ := runTool(t, tools, "nuage_search", `{"query":"notes"}`); !strings.Contains(hits, `"path":"/notes.txt"`) {
		t.Errorf("nuage_search: %s", hits)
	}
	if text, _ := runTool(t, tools, "nuage_read", `{"file_id":1}`); !strings.Contains(text, `"content":"hello nuage"`) {
		t.Errorf("nuage_read text: %s", text)
	}
	image, _ := runTool(t, tools, "nuage_read", `{"file_id":2}`)
	if strings.Contains(image, "content") || !strings.Contains(image, "attach the file") {
		t.Errorf("nuage_read image: %s", image)
	}
}

func TestFilesSaveFromContentAndAttachment(t *testing.T) {
	w, rt := nuageWorld(t)
	tools, _, _ := Tools(NewGateway(w.tokens(), w.cfg, nil), 1, Defaults{}, rt)
	if _, err := runTool(t, tools, "nuage_save", `{"name":"todo.md","content":"- milk","folder":3}`); err != nil {
		t.Fatal(err)
	}
	if _, err := runTool(t, tools, "nuage_save", `{"attachment_id":"att-1"}`); err != nil {
		t.Fatal(err)
	}
	saved := w.fakes.saved
	if len(saved) != 2 || saved[0].Name != "todo.md" || *saved[0].FolderID != 3 || saved[1].Name != "scan.pdf" || saved[1].MimeType != "application/pdf" {
		t.Fatalf("saved: %+v", saved)
	}
	if string(w.fakes.uploads[saved[0].ID]) != "- milk" || string(w.fakes.uploads[saved[1].ID]) != "%PDF-1.7" {
		t.Fatalf("uploaded bytes differ")
	}
	for _, input := range []string{`{"name":"a.md"}`, `{"name":"a.md","content":"x","attachment_id":"att-1"}`, `{"name":"../a.md","content":"x"}`, `{"attachment_id":"nope"}`} {
		if _, err := runTool(t, tools, "nuage_save", input); err == nil {
			t.Errorf("%s was accepted", input)
		}
	}
}
