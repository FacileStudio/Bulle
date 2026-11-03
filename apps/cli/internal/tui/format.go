package tui

import (
	"fmt"
	"strings"

	"github.com/FacileStudio/nacelle"
)

// question draws the reader's question with the left half-block spine (▌) on
// every row, wrapped or not: the body renders at two columns short of the
// width so each line the style breaks has room for the prefix, and the prefix
// is added after rendering so wrapped rows carry it too.
func (m *Model) question(text string, width int) string {
	body := m.theme.Question.Width(max(width-2, 1)).Render(text)
	spines := m.theme.Question.Render("▌ ")
	tokens := m.skillTokens()
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		lines[i] = spines + highlightSkills(line, tokens)
	}
	return strings.Join(lines, "\n")
}

// countedNoun formats a count with its English plural, singular for exactly
// one.
func countedNoun(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

func cutShort(stop nacelle.Stop) string {
	if stop == "" || stop.Complete() {
		return ""
	}
	switch stop {
	case nacelle.StopMaxTokens:
		return "cut off at the token limit"
	case nacelle.StopContext:
		return "cut off: out of context"
	case nacelle.StopRefusal:
		return "refused by the model"
	case nacelle.StopIterations:
		return "stopped at the iteration limit"
	case abandoned:
		return "abandoned"
	}
	return "stopped early"
}
