package tui

import (
	"image/color"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/FacileStudio/Bulle/apps/cli/internal/theme"
)

// A light-background report repaints the held transcript in the light
// palette, rather than leaving the dark rows until the next resize.
func TestRethemeRepaintsTheHold(t *testing.T) {
	m := tuiModel()
	defer theme.Themed(true)
	m.say(fromReader, "what changed")
	m.prints()
	dark := strings.Join(m.hold, "\n")

	m.retheme(tea.BackgroundColorMsg{Color: color.White})

	if strings.Join(m.hold, "\n") == dark {
		t.Error("the hold kept its dark paint after a light retheme")
	}
}
