package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/FacileStudio/Bulle/apps/cli/internal/ask"
	"github.com/FacileStudio/Bulle/core/askuser"
)

var dayAndTone = []askuser.Question{
	{Question: "Which day?", Options: []askuser.Option{{Label: "Monday"}, {Label: "Tuesday"}}},
	{Question: "Tone?", MultiSelect: true, Options: []askuser.Option{{Label: "Warm"}, {Label: "Dry"}, {Label: "Short"}}},
}

func asked(m *Model) chan []askuser.Answer {
	answers := make(chan []askuser.Answer, 1)
	m.Update(ask.Request{Answers: answers, Questions: dayAndTone})
	return answers
}

func TestQuestionTakesDigitsTogglesAndTypedAnswers(t *testing.T) {
	m := bareBanner()
	answers := asked(m)
	if !strings.Contains(m.questionView(), "Which day?") {
		t.Fatalf("question not drawn: %q", m.questionView())
	}
	m.Update(tea.KeyPressMsg{Code: '2', Text: "2"})
	m.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	m.Update(tea.KeyPressMsg{Code: '3', Text: "3"})
	m.promptRoute(tea.PasteMsg{Content: "no jokes"})
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got := <-answers
	if m.run.question != nil {
		t.Fatal("the question is still pending")
	}
	if reply := askuser.Reply(dayAndTone, got); reply != "Which day? Tuesday\nTone? Warm, Short, no jokes" {
		t.Fatalf("reply %q", reply)
	}
}

func TestQuestionEscDismisses(t *testing.T) {
	m := bareBanner()
	answers := asked(m)
	m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if got := <-answers; got != nil || m.run.question != nil {
		t.Fatalf("esc answered %+v", got)
	}
}
