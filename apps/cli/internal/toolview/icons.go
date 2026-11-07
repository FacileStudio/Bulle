package toolview

import (
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"

	"github.com/FacileStudio/nacelle"

	"github.com/FacileStudio/Bulle/apps/cli/internal/theme"
)

var toolIcons = map[string]string{
	"bash":            "$",
	"edit_file":       "⏺",
	"write_file":      "⏺",
	"read_file":       "⏺",
	"search_files":    "⏺",
	"glob":            "⏺",
	"grep":            "⏺",
	"grep_files":      "⏺",
	"web_fetch":       "⏺",
	"download":        "⏺",
	"parallel_agents": "⏺",
}

var toolKinds = map[string]string{
	"bash":            "write",
	"edit_file":       "write",
	"write_file":      "write",
	"read_file":       "read",
	"search_files":    "read",
	"glob":            "read",
	"grep":            "read",
	"grep_files":      "read",
	"web_fetch":       "network",
	"download":        "network",
	"parallel_agents": "delegate",
}

// ToolKind categorizes a tool by its operation group for batching.
func ToolKind(name string, source nacelle.Source) string {
	if source == nacelle.ToolSourceMCP {
		return "mcp"
	}
	if kind, ok := toolKinds[name]; ok {
		return kind
	}
	return "other"
}

// ToolKindGlyph returns the representative icon for a tool kind.
func ToolKindGlyph(kind string) string {
	switch kind {
	case "mcp":
		return "✻"
	default:
		return "⏺"
	}
}

// ToolGlyph returns the marker icon for a named tool.
func ToolGlyph(name string) string {
	if icon, ok := toolIcons[name]; ok {
		return icon
	}
	return "⏺"
}

// ToolTone returns the style of a named tool's line: every tool reads in
// the same muted grey, and only the glyph carries the call's state.
func ToolTone(string) lipgloss.Style {
	return theme.PlainTool()
}

// ToolRestore returns the escape parameters that bring a line back to its
// tool tone after a coloured glyph.
func ToolRestore(string) string {
	return theme.SGR(theme.Ink.Muted)
}

// ToolLinePainted paints a held call line in the tool tone.
func ToolLinePainted(text string) string {
	return theme.PlainTool().Render(text)
}

// ColorGlyph colours the initial glyph of a line with color and restores restoreColour.
func ColorGlyph(line, color, restoreColour string) string {
	for i := 0; i < len(line); i++ {
		if line[i] == '\x1b' {
			end := strings.IndexByte(line[i:], 'm')
			if end < 0 {
				break
			}
			i += end
			continue
		}
		r, size := utf8.DecodeRuneInString(line[i:])
		if r == utf8.RuneError {
			break
		}
		before := line[:i]
		glyph := line[i : i+size]
		rest := line[i+size:]
		return before + "\x1b[" + color + "m" + glyph + "\x1b[" + restoreColour + "m" + rest
	}
	return line
}
