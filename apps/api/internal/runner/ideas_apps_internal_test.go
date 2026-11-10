package runner

import (
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

func TestEverySuiteAppHasItsOwnIdeasGuidance(t *testing.T) {
	for _, app := range store.IdeaApps() {
		if app == store.IdeaAppBulle {
			continue
		}
		guide, ok := appGuides[app]
		if !ok || !strings.Contains(guide, "in this order") || !strings.Contains(guide, "Use kind") {
			t.Errorf("%s has no ordered guidance with kinds: %q", app, guide)
		}
		if !strings.Contains(appGuidance(app), "Start every fingerprint with "+app+":") {
			t.Errorf("%s guidance does not key its fingerprints", app)
		}
	}
}
