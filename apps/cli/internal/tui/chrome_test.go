package tui

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/cli/internal/approval"
	"github.com/FacileStudio/Bulle/apps/cli/internal/toolview"
	"github.com/FacileStudio/nacelle"
)

// Durations read one decimal of seconds past a second.
func TestTookShowsOneDecimalAboveASecond(t *testing.T) {
	cases := map[time.Duration]string{
		450 * time.Millisecond:  "450ms",
		2347 * time.Millisecond: "2.3s",
		65 * time.Second:        "1m5s",
	}
	for d, want := range cases {
		if got := toolview.Took(d); got != want {
			t.Errorf("Took(%v) = %q, want %q", d, got, want)
		}
	}
}

// No ↑0 ↓0 before anything was billed.
func TestFooterHidesZeroTokensBeforeTheFirstTurn(t *testing.T) {
	m := sized()
	if foot := strings.Join(m.footer(), " "); strings.Contains(foot, "↑0") {
		t.Errorf("footer = %q, want no zero token counts", foot)
	}
}

// A refused call reads "denied", not a failure carrying nacelle's error.
func TestADeniedCallReadsDenied(t *testing.T) {
	m := sized()
	call := &nacelle.ToolEvent{ID: "d1", Name: "bash", Input: `{"command":"rm -rf /"}`}
	m.absorb(nacelle.Event{Kind: nacelle.KindToolCall, Tool: call})
	refused := *call
	refused.Err = errors.New(`nacelle: "bash" was refused: the user did not approve it`)
	m.absorb(nacelle.Event{Kind: nacelle.KindToolResult, Tool: &refused})
	m.flushFailures()
	said := visible(strings.Join(m.unprinted, "\n"))
	if !strings.Contains(said, "denied") || strings.Contains(said, "not approved") || m.failed != 0 {
		t.Errorf("said %q with %d failures, want a quiet denied line", said, m.failed)
	}
}

// While its approval is shown, the running call reads "waiting for approval".
func TestRunningCallWaitsForApproval(t *testing.T) {
	m := sized()
	input := `{"pattern":"x"}`
	m.run.beginTool(nacelle.ToolEvent{ID: "s1", Name: "search", Input: input}, false)
	m.parkApproval(approval.Request{Name: "search", Input: []byte(input), Decision: make(chan approval.Decision, 1)})
	if rows := visible(strings.Join(m.inFlightGroups(), "\n")); !strings.Contains(rows, "waiting for approval") {
		t.Errorf("in-flight rows = %q, want the waiting note", rows)
	}
}

// /clear in tui mode drops the painted hold and the scroll offset.
func TestClearResetsTheHoldAndScroll(t *testing.T) {
	m := tuiModel()
	m.hold = []string{"old", "rows"}
	m.scrollTop = 4
	m.clear()
	if m.scrollTop != 0 || strings.Contains(strings.Join(m.hold, "\n"), "old") {
		t.Errorf("hold %q scrollTop %d survived /clear", m.hold, m.scrollTop)
	}
}
