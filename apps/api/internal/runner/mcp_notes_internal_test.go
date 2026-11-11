package runner

import (
	"slices"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
)

func TestEveryMCPNoteIsKeyedOnAServedTool(t *testing.T) {
	served := slices.Concat(suite.CourrierTools(), suite.AgendaTools(), suite.SonarTools(), suite.NuageTools(),
		suite.OpusTools(), suite.SablierTools(), suite.GloutonTools(), suite.PlumeTools(), suite.CapsuleTools(), suite.JournalTools(),
		suite.AntenneTools(), suite.VisionTools(), suite.EchoTools())
	for _, section := range mcpSections {
		if !slices.Contains(served, section.tool) {
			t.Errorf("the note keyed on %s never shows: no server exposes it", section.tool)
		}
	}
	fresh := []string{"plume_", "capsule_", "journal_", "antenne_", "vision_", "echo_"}
	for _, name := range untrustedSources {
		app := slices.IndexFunc(fresh, func(prefix string) bool { return strings.HasPrefix(name, prefix) })
		if app >= 0 && !slices.Contains(served, name) {
			t.Errorf("%s taints a run but no server exposes it", name)
		}
	}
}
