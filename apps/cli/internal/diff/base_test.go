package diff

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// An edit_file diff numbers its rows as the file does: the old text sits on
// line 41, so its removal reads 41, not 1.
func TestEditDiffNumbersRowsAsTheFileDoes(t *testing.T) {
	root := t.TempDir()
	var file strings.Builder
	for range 40 {
		file.WriteString("filler\n")
	}
	file.WriteString("target\nafter\n")
	if err := os.WriteFile(filepath.Join(root, "f.txt"), []byte(file.String()), 0o600); err != nil {
		t.Fatal(err)
	}
	change, ok := CaptureEdit(root, "edit_file", `{"path":"f.txt","old":"target","new":"hit"}`)
	if !ok || change.Base != 40 {
		t.Fatalf("change = %+v, %v; want Base 40", change, ok)
	}
	if text := plain(RenderDiff(change, 80, "32", muted, false)); !strings.Contains(text, "41") {
		t.Errorf("diff does not number the row 41:\n%s", text)
	}
}
