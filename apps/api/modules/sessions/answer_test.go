package sessions

import (
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
)

func TestTheOpenPageNoteStaysOutOfTheBubble(t *testing.T) {
	note, err := runner.PageContext{App: "docs", ID: "p1", Title: "Plan", URL: "https://docs.facile.studio/p/p1"}.Note()
	if err != nil {
		t.Fatal(err)
	}
	got := userMessages([]runner.WirePart{{Type: "text", Text: note.Text}, {Type: "text", Text: "what next?"}})
	if len(got) != 1 || got[0].Content != "what next?" {
		t.Fatalf("bubbles = %+v", got)
	}
}
