package tui

import (
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/cli/internal/approval"
)

// The approval shows the whole command, wrapped, rather than the 60 characters
// the status line used to fit, and the live region gives up the rows it takes.
func TestApprovalShowsTheFullCommandAndTakesItsRows(t *testing.T) {
	m := sized()
	before := m.liveRows
	command := "rm -rf ./build && go build -o ./build/bulle ./apps/cli && ./build/bulle --version --verbose && echo finished-the-tail"
	m.parkApproval(approval.Request{Name: "run_command", Input: []byte(`{"command":"` + command + `"}`), Decision: make(chan approval.Decision, 1)})

	view := visible(m.approvalView())
	flat := strings.Join(strings.Fields(strings.ReplaceAll(view, "▌", " ")), " ")
	if !strings.Contains(flat, "finished-the-tail") {
		t.Fatalf("approval view lost the end of the command:\n%s", view)
	}
	if m.liveRows != before-m.approvalRows() {
		t.Errorf("liveRows = %d, want %d less the approval's %d rows", m.liveRows, before, m.approvalRows())
	}
	if strings.Contains(visible(m.status()), "rm -rf") {
		t.Errorf("status still carries the input: %q", m.status())
	}
}

// An edit's approval previews the diff it will make.
func TestApprovalPreviewsAnEditAsADiff(t *testing.T) {
	m := sized()
	input := `{"path":"notes.txt","old":"alpha\nbeta","new":"alpha\ngamma"}`
	m.parkApproval(approval.Request{Name: "edit_file", Input: []byte(input), Decision: make(chan approval.Decision, 1)})

	view := visible(m.approvalView())
	for _, want := range []string{"notes.txt", "gamma", "beta"} {
		if !strings.Contains(view, want) {
			t.Errorf("diff preview missing %q:\n%s", want, view)
		}
	}
}
