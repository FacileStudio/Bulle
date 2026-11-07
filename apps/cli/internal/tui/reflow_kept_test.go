package tui

import (
	"errors"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/FacileStudio/Bulle/apps/cli/internal/toolview"
	"github.com/FacileStudio/nacelle"
)

// Failures and the fan-out notice used to go straight to the unprinted queue,
// so a resize, which repaints the hold from the held entries, lost them.
func TestResizeKeepsFailuresAndNotices(t *testing.T) {
	m := tuiModel()
	m.announceStart(2)
	tool := nacelle.ToolEvent{ID: "t1", Name: "read_file", Err: errors.New("no such file"), Duration: time.Second}
	m.trackFailure("read_file missing.txt", &tool)
	m.flushFailures()
	m.printGroupFailure("search x", "search", []toolview.ToolError{{Name: "search", Err: "bad pattern"}}, "1s")
	m.prints()

	m.resize(tea.WindowSizeMsg{Width: 50, Height: 24})

	screen := visible(strings.Join(m.hold, "\n"))
	for _, want := range []string{"2 parallel agents started", "no such file", "bad pattern"} {
		if !strings.Contains(screen, want) {
			t.Errorf("resize lost %q:\n%s", want, screen)
		}
	}
}
