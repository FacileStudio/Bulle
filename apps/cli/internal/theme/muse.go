package theme

import (
	"fmt"
	"image/color"
	"strconv"

	"charm.land/glamour/v2/ansi"
	"charm.land/lipgloss/v2"
)

// Muse is the muse design system's palette as hex, the same tokens the web
// client paints with (bg-fc-surface, text-fc-fg-muted, ...), so the terminal
// and the browser read as one product. Colour carries state only: danger,
// warning, success and info; everything else is a grey.
type Muse struct {
	Surface   string
	Component string
	Fg        string
	Muted     string
	Dim       string
	Danger    string
	Success   string
	Warning   string
	Info      string
	AddBg     string
	DelBg     string
}

var (
	museDark = Muse{
		Surface: "#181818", Component: "#0d0d0d", Fg: "#fafafa", Muted: "#8c8c8c", Dim: "#3d3d3d",
		Danger: "#f94335", Success: "#5bbd74", Warning: "#eeb154", Info: "#6aa7f4",
		AddBg: "#12261a", DelBg: "#2a1414",
	}
	museLight = Muse{
		Surface: "#eeeeee", Component: "#f7f7f7", Fg: "#0a0a0a", Muted: "#636363", Dim: "#c8c8c8",
		Danger: "#d00405", Success: "#21763c", Warning: "#935b00", Info: "#2366b4",
		AddBg: "#e6f4ea", DelBg: "#fbe9e9",
	}
)

// Ink is the palette of the terminal the program draws on, dark until the
// terminal reports a light background. Every renderer outside the model reads
// it, so SetDark, called from the one place that learns the background, is the
// only writer.
var Ink = museDark

// SetDark picks the dark or light palette.
func SetDark(dark bool) {
	Ink = museLight
	if dark {
		Ink = museDark
	}
}

// C is a palette hex as a lipgloss colour.
func C(hex string) color.Color { return lipgloss.Color(hex) }

// Fore is a foreground style in a palette colour.
func Fore(hex string) lipgloss.Style { return lipgloss.NewStyle().Foreground(C(hex)) }

// SGR is the escape parameters that set a palette hex as the foreground,
// for the renderers that splice raw escapes into a line.
func SGR(hex string) string {
	if len(hex) != 7 {
		return "39"
	}
	value, err := strconv.ParseUint(hex[1:], 16, 32)
	if err != nil {
		return "39"
	}
	return fmt.Sprintf("38;2;%d;%d;%d", value>>16&0xff, value>>8&0xff, value&0xff)
}

// museMarkdown restyles glamour's layout in muse greys: text in the
// foreground, inline code on the surface instead of glamour's red, links in
// info, headings bold and uncoloured.
func museMarkdown(base ansi.StyleConfig, ink Muse) ansi.StyleConfig {
	bold, fg, muted, surface, info := true, ink.Fg, ink.Muted, ink.Surface, ink.Info
	base.Document.Color = &fg
	base.Text.Color = &fg
	base.Paragraph.Color = &fg
	base.Heading.Color, base.Heading.Bold = &fg, &bold
	for _, h := range []*ansi.StyleBlock{&base.H1, &base.H2, &base.H3, &base.H4, &base.H5, &base.H6} {
		h.Color, h.BackgroundColor = &fg, nil
	}
	base.Code.Color, base.Code.BackgroundColor = &fg, &surface
	base.CodeBlock.Color = &fg
	base.Link.Color, base.LinkText.Color = &info, &info
	base.BlockQuote.Color = &muted
	base.HorizontalRule.Color = &muted
	base.Item.Color, base.Enumeration.Color = &muted, &muted
	return base
}
