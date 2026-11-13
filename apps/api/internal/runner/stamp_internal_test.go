package runner

import (
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/nacelle"
)

func TestStampPromptDatesTheStoredPromptOnce(t *testing.T) {
	now := time.Date(2026, 11, 3, 14, 5, 0, 0, time.FixedZone("CET", 3600))
	stamped := stampPrompt(nacelle.UserText("what day is it?"), now)
	clock, ok := stamped.Parts[0].(nacelle.Text)
	if !ok || !strings.Contains(clock.Text, "Tuesday 3 November 2026, 14:05 CET") || !isStamp(clock.Text) {
		t.Fatalf("the prompt is not stamped: %#v", stamped.Parts)
	}
	if len(stamped.Parts) != 2 || isStamp("what day is it?") {
		t.Fatal("the stamp replaced the prompt or matched plain text")
	}
}
