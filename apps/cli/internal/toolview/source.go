package toolview

import (
	"charm.land/lipgloss/v2"

	"github.com/FacileStudio/Bulle/apps/cli/internal/theme"
	"github.com/FacileStudio/nacelle"
)

// ToolSourceGlyph returns the marker icon for a tool by its nacelle source.
// It is source-first: MCP tools get ✻ regardless of name. Everything else
// falls back to the normal name-based glyph.
func ToolSourceGlyph(name string, source nacelle.Source) string {
	if source == nacelle.ToolSourceMCP {
		return "✻"
	}
	return ToolGlyph(name)
}

// ToolSourceTone returns the Lipgloss style for a tool by its nacelle source.
func ToolSourceTone(name string, source nacelle.Source) lipgloss.Style {
	if source == nacelle.ToolSourceMCP {
		return theme.Fore(theme.Ink.Info)
	}
	return ToolTone(name)
}

// ToolSourceRestore returns the ANSI colour code for a tool by its nacelle source.
func ToolSourceRestore(name string, source nacelle.Source) string {
	if source == nacelle.ToolSourceMCP {
		return theme.SGR(theme.Ink.Info)
	}
	return ToolRestore(name)
}

// ToolSourceColor returns the escape parameters for a tool's leading glyph:
// success or danger for the call's outcome, info for an MCP tool that worked.
func ToolSourceColor(_ string, source nacelle.Source, ok bool) string {
	switch {
	case !ok:
		return theme.SGR(theme.Ink.Danger)
	case source == nacelle.ToolSourceMCP:
		return theme.SGR(theme.Ink.Info)
	}
	return theme.SGR(theme.Ink.Success)
}

// ToolBorder returns the spine colour of a box whose tool is still running:
// the warning tone, for every tool alike. A finished box trades it for the
// outcome's colour via the caller.
func ToolBorder() string {
	return theme.Ink.Warning
}

// ToolLineRunning paints a held tool line in the warning tone while its call
// is still running, the colour its box spine wears underneath.
func ToolLineRunning(text string) string {
	return theme.Fore(theme.Ink.Warning).Render(text)
}
