package tui

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/FacileStudio/blob"

	"github.com/FacileStudio/Bulle/apps/cli/internal/layout"
)

const (
	defaultBarLabel = "Bulle"
	barLabelMax     = 16
	loaderWidth     = 3
)

// barLabel is the name the prompt shows on its first row and the colour its
// loader takes: the signed-in agent's name and the head colour of its blob, so
// the loader matches the avatar the session opened on. Without an agent it is
// Bulle on the bar's own greys.
type barLabel struct {
	name   string
	accent color.Color
}

func newBarLabel(agent string) barLabel {
	if agent == "" {
		return barLabel{name: defaultBarLabel}
	}
	return barLabel{
		name:   layout.Truncate(agent, barLabelMax),
		accent: lipgloss.Color(blob.New(agent, blob.Options{}).Palette.Head.Hex()),
	}
}

func (l barLabel) width() int {
	return max(lipgloss.Width(l.name), loaderWidth)
}

func (l barLabel) gutterWidth() int {
	return barPad + l.width() + barPad
}

// loaderShades is the lit box and the one trailing it: the accent and the
// accent half faded into the panel, or the bar's text and muted greys.
func (l barLabel) loaderShades(panel, text, muted color.Color) (color.Color, color.Color) {
	if l.accent == nil {
		return text, muted
	}
	return l.accent, lipgloss.Blend1D(3, panel, l.accent)[1]
}
