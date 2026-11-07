package tui

import (
	"strings"

	"github.com/FacileStudio/Bulle/apps/cli/internal/menu"
)

// viewMenu is the slash menu, inset by the same one-column margin as the rest
// of the live region.
func (m *Model) viewMenu() string {
	view := menu.View(&m.menu, max(m.width-2, 1), m.theme.Plain, m.theme.Menu, m.theme.Command)
	if view == "" {
		return ""
	}
	return strings.Join(m.margined([]string{view}), "\n")
}
