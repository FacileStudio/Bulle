package toolview

import (
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/cli/internal/theme"
	"github.com/FacileStudio/nacelle"
)

func TestMCPSourceGlyph(t *testing.T) {
	if got := ToolSourceGlyph("whatever_mcp_tool", nacelle.ToolSourceMCP); got != "✻" {
		t.Errorf("ToolSourceGlyph(MCP) = %q, want ✻", got)
	}
	if got := ToolSourceGlyph("bash", nacelle.ToolSourceLocal); got != "$" {
		t.Errorf("ToolSourceGlyph(local) = %q, want $", got)
	}
}

func TestMCPSourceRestore(t *testing.T) {
	if got := ToolSourceRestore("whatever_mcp_tool", nacelle.ToolSourceMCP); got != theme.SGR(theme.Ink.Info) {
		t.Errorf("ToolSourceRestore(MCP) = %q, want the info tone", got)
	}
	if got := ToolSourceRestore("_", nacelle.ToolSourceLocal); got != theme.SGR(theme.Ink.Muted) {
		t.Errorf("ToolSourceRestore(unknown local) = %q, want the muted tone", got)
	}
}

func TestMCPSourceColor(t *testing.T) {
	if got := ToolSourceColor("whatever_mcp_tool", nacelle.ToolSourceMCP, true); got != theme.SGR(theme.Ink.Info) {
		t.Errorf("ToolSourceColor(MCP ok) = %q, want the info tone", got)
	}
	for _, source := range []nacelle.Source{nacelle.ToolSourceMCP, nacelle.ToolSourceLocal} {
		if got := ToolSourceColor("bash", source, false); got != theme.SGR(theme.Ink.Danger) {
			t.Errorf("ToolSourceColor(%s fail) = %q, want the danger tone", source, got)
		}
	}
	if got := ToolSourceColor("bash", nacelle.ToolSourceLocal, true); got != theme.SGR(theme.Ink.Success) {
		t.Errorf("ToolSourceColor(local ok) = %q, want the success tone", got)
	}
}

// ToolBorder returns the running state's one colour — yellow, basic "3" — for
// every tool, local or MCP alike, so a running box's spine always reads as
// "in flight". The finished box trades this for green or red via the caller.
func TestToolBorderIsAlwaysYellowWhileRunning(t *testing.T) {
	if got := ToolBorder(); got != theme.Ink.Warning {
		t.Errorf("ToolBorder() = %q, want the warning tone", got)
	}
}

// ToolLineRunning paints a held line yellow, the same running-state colour the
// box spine wears, so a live row reads as "in flight" whatever its tool tone.
func TestToolLineRunningPaintsYellow(t *testing.T) {
	if got := ToolLineRunning("✎ edit_file(view.go)"); !strings.Contains(got, "\x1b["+theme.SGR(theme.Ink.Warning)) {
		t.Errorf("ToolLineRunning = %q, want the warning tone", got)
	}
}
