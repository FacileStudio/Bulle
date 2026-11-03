package tui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/textarea"
	"charm.land/lipgloss/v2"
	"github.com/FacileStudio/blob"
)

func TestPromptLabelNamesTheAgentAndLoaderTakesItsBlobColour(t *testing.T) {
	m := sized()
	m.label = newBarLabel("Jarvis")
	if got := visible(m.gutter(textarea.PromptInfo{})); got != "  Jarvis  " {
		t.Errorf("gutter = %q, want the agent name", got)
	}
	m.run.busy = true
	if got := visible(m.gutter(textarea.PromptInfo{})); len([]rune(got)) != m.label.gutterWidth() {
		t.Errorf("busy gutter = %q, want the loader padded to the label width", got)
	}
	lit, _ := m.label.loaderShades(m.theme.Bar.Panel, m.theme.Bar.Text, m.theme.Bar.Muted)
	if want := lipgloss.Color(blob.New("Jarvis", blob.Options{}).Palette.Head.Hex()); lit != want {
		t.Errorf("loader colour = %v, want the blob head %v", lit, want)
	}
	if got := newBarLabel("").name; got != "Bulle" {
		t.Errorf("label without agent = %q, want Bulle", got)
	}
}

func TestFooterStartsAtTheLeftEdge(t *testing.T) {
	m := sized()
	m.activeModel = "claude-haiku"
	if foot := visible(m.barFooter()); !strings.HasPrefix(foot, "claude-haiku") {
		t.Errorf("footer = %q, want the model at column 0", foot)
	}
}
