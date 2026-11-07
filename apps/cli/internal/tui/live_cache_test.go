package tui

import (
	"strings"
	"testing"
	"time"
)

// An unchanged key never repaints; a text change inside the gap keeps the old
// render; a width change repaints at once.
func TestCachedRenderPaintsOnlyWhenItMust(t *testing.T) {
	m := sized()
	m.live.gap = time.Hour
	var slot cachedPaint
	calls := 0
	paint := func() string { calls++; return "painted" }

	m.cachedRender(&slot, m.paintKey("a", 80), paint)
	m.cachedRender(&slot, m.paintKey("a", 80), paint)
	if calls != 1 {
		t.Fatalf("same key painted %d times, want 1", calls)
	}
	m.cachedRender(&slot, m.paintKey("ab", 80), paint)
	if calls != 1 {
		t.Fatalf("a text change inside the gap repainted")
	}
	m.cachedRender(&slot, m.paintKey("ab", 60), paint)
	if calls != 2 {
		t.Fatalf("a width change did not repaint")
	}
}

// A long open fence renders as its header and its last lines only; a closed
// fence and plain prose pass through whole.
func TestFenceTailKeepsTheHeaderAndTheTail(t *testing.T) {
	var body strings.Builder
	for i := range 100 {
		body.WriteString("line" + string(rune('A'+i%26)) + "\n")
	}
	text := "intro\n```go\n" + body.String() + "last"
	got := fenceTail(text, 5)
	if !strings.HasPrefix(got, "```go\n") || !strings.HasSuffix(got, "last") {
		t.Fatalf("fenceTail lost the header or the tail: %q", got)
	}
	if n := strings.Count(got, "\n"); n != 5 {
		t.Errorf("fenceTail kept %d lines after the header, want 5", n)
	}
	closed := text + "\n```\nafter"
	if fenceTail(closed, 5) != closed {
		t.Error("a closed fence was cut")
	}
}

// The streaming answer pane is rendered once for an unchanged answer, so a
// spinner tick costs no markdown render.
func TestStreamingReusesTheAnswerRender(t *testing.T) {
	m := sized()
	m.run.answer.WriteString("hello **world**")
	m.streaming()
	first := m.live.answer.at
	m.streaming()
	if m.live.answer.at != first {
		t.Error("an unchanged answer was rendered again")
	}
}
