// Package theme defines styles and color palettes for the terminal UI.
package theme

import (
	"image/color"
	"regexp"
	"strings"

	"charm.land/glamour/v2"
	"charm.land/glamour/v2/styles"
	"charm.land/lipgloss/v2"
)

var osc8 = regexp.MustCompile(`\x1b\]8;[^\x1b\x07]*(?:\x07|\x1b\\)`)

// PlainTool returns the style of a tool line: a muted grey, so a run of
// calls reads as steps under the conversation rather than part of it.
func PlainTool() lipgloss.Style {
	return Fore(Ink.Muted)
}

// TranscriptStyles contains styles used when formatting transcript rows.
type TranscriptStyles struct {
	Question lipgloss.Style
	Thinking lipgloss.Style
	Tool     lipgloss.Style
	Result   lipgloss.Style
	Failure  lipgloss.Style
	Client   lipgloss.Style
	Command  lipgloss.Style
}

// UIStyles contains styles used for UI elements outside the transcript.
type UIStyles struct {
	Menu       lipgloss.Style
	Plain      lipgloss.Style
	Waiting    lipgloss.Style
	Ready      lipgloss.Style
	Muted      lipgloss.Style
	Border     lipgloss.Style
	Compacting lipgloss.Style
}

// Palette groups all styles and settings needed to render the interface.
type Palette struct {
	TranscriptStyles
	UIStyles
	Markdown string
	Bar      Bar
}

// Bar is the input panel and the line under it, after grok-cli's prompt: a
// flat panel, text, then muted and dim greys for labels and figures.
type Bar struct {
	Panel color.Color
	Text  color.Color
	Muted color.Color
	Dim   color.Color
}

func barFor(ink Muse) Bar {
	return Bar{Panel: C(ink.Surface), Text: C(ink.Fg), Muted: C(ink.Muted), Dim: C(ink.Dim)}
}

func transcriptStylesFor(ink Muse) TranscriptStyles {
	return TranscriptStyles{
		Question: lipgloss.NewStyle().Background(C(ink.Surface)).Foreground(C(ink.Fg)),
		Thinking: Fore(ink.Muted).Italic(true),
		Tool:     PlainTool(),
		Result:   Fore(ink.Muted),
		Failure:  Fore(ink.Danger),
		Client:   Fore(ink.Muted),
		Command:  Fore(ink.Info),
	}
}

func uiStylesFor(ink Muse) UIStyles {
	return UIStyles{
		Menu:       lipgloss.NewStyle().Background(C(ink.Surface)).Foreground(C(ink.Fg)),
		Plain:      lipgloss.NewStyle(),
		Waiting:    Fore(ink.Info),
		Ready:      Fore(ink.Success),
		Muted:      Fore(ink.Muted),
		Border:     Fore(ink.Dim),
		Compacting: Fore(ink.Info),
	}
}

// Themed builds the palette for a light or dark terminal and makes it the
// Ink every renderer outside the model reads.
func Themed(dark bool) Palette {
	SetDark(dark)
	style := "light"
	if dark {
		style = "dark"
	}
	return Palette{
		TranscriptStyles: transcriptStylesFor(Ink),
		UIStyles:         uiStylesFor(Ink),
		Markdown:         style,
		Bar:              barFor(Ink),
	}
}

// Prettier builds the glamour markdown renderer.
//
// The layout is glamour's dark or light one, recoloured in muse greys by
// museMarkdown. Width is the word-wrap width.
func Prettier(style string, width int) *glamour.TermRenderer {
	base, ink := styles.LightStyleConfig, museLight
	if style == "dark" {
		base, ink = styles.DarkStyleConfig, museDark
	}
	base = museMarkdown(base, ink)
	base.Document.Margin = nil
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(base),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil
	}
	return renderer
}

// RenderMarkdown formats markdown text for terminal display.
//
// Glamour frames every render with a top-margin newline and trailing fill to
// the wrap width. Answers stream into the terminal as committed fragments (see
// commitParagraphs), each rendered on its own and joined back together in
// prints, so that framing would otherwise leave a blank line between every
// fragment. Trim the framing here and keep the paragraph's own indent.
func RenderMarkdown(r *glamour.TermRenderer, text string) string {
	if r == nil {
		return text
	}
	rendered, err := r.Render(text)
	if err != nil {
		return text
	}
	rendered = dropHyperlinks(rendered)
	return strings.TrimLeft(strings.TrimRight(rendered, " \n"), "\n")
}

// dropHyperlinks removes OSC 8 hyperlink sequences.
//
// tmux does not speak OSC 8 and forwards the sequences mangled, which leaves
// kitty — the terminal behind tmux over ssh — with a hyperlink open past its
// reset: everything after renders underlined. Plain text keeps the link label
// and reads the same.
func dropHyperlinks(rendered string) string {
	return osc8.ReplaceAllString(rendered, "")
}
