package tui

import "github.com/FacileStudio/Bulle/apps/cli/internal/menu"

func (m *Model) viewMenu() string {
	return menu.View(&m.menu, max(m.width, 1), m.theme.Plain, m.theme.Menu, m.theme.Command)
}
