package tui

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/FacileStudio/Bulle/apps/cli/internal/ask"
	"github.com/FacileStudio/Bulle/apps/cli/internal/herdr"
	"github.com/FacileStudio/Bulle/core/askuser"
)

// question is an ask_user call waiting for the reader: which question is
// shown, which option the cursor is on, and the answers given so far.
type question struct {
	req     ask.Request
	index   int
	cursor  int
	answers []askuser.Answer
}

func (q *question) current() askuser.Question { return q.req.Questions[q.index] }

// parkQuestion holds an ask_user call until the reader answers every
// question or dismisses them.
func (m *Model) parkQuestion(req ask.Request) tea.Cmd {
	m.dropDeadQuestion()
	m.run.question = &question{req: req, answers: make([]askuser.Answer, len(req.Questions))}
	herdr.Report(m.herdrClient, herdr.Blocked)
	m.layout(m.windowHeight)
	return nil
}

// questionKey answers the keys a pending question owns: arrows move, a digit
// or enter picks, space toggles a multi-choice option, esc dismisses. Digits
// and space belong to the prompt once the reader types their own answer, and
// every other key falls through to it.
func (m *Model) questionKey(press tea.KeyPressMsg) bool {
	q := m.run.question
	empty := strings.TrimSpace(m.prompt.Value()) == ""
	key := press.String()
	switch {
	case key == "esc":
		m.finishQuestion(nil)
	case key == "up" || key == "ctrl+k":
		q.cursor = (q.cursor + len(q.current().Options) - 1) % len(q.current().Options)
	case key == "down" || key == "ctrl+j":
		q.cursor = (q.cursor + 1) % len(q.current().Options)
	case key == "enter":
		m.confirmQuestion()
	case empty && key == "space" && q.current().MultiSelect:
		q.toggle(q.cursor)
	case empty && len(key) == 1 && key >= "1" && key <= "4":
		m.pickNumber(int(key[0] - '1'))
	default:
		return false
	}
	return true
}

func (m *Model) pickNumber(n int) {
	q := m.run.question
	if n >= len(q.current().Options) {
		return
	}
	q.cursor = n
	if q.current().MultiSelect {
		q.toggle(n)
		return
	}
	m.confirmQuestion()
}

func (q *question) toggle(i int) {
	label := q.current().Options[i].Label
	answer := &q.answers[q.index]
	if at := slices.Index(answer.Picked, label); at >= 0 {
		answer.Picked = slices.Delete(answer.Picked, at, at+1)
		return
	}
	answer.Picked = append(answer.Picked, label)
}

// confirmQuestion takes the typed answer when there is one, else the option
// under the cursor (for a multi-choice question, only when nothing is
// toggled), then moves on to the next question.
func (m *Model) confirmQuestion() {
	q := m.run.question
	answer := &q.answers[q.index]
	if typed := strings.TrimSpace(m.prompt.Value()); typed != "" {
		answer.Typed = typed
		m.prompt.Reset()
	} else if !q.current().MultiSelect || len(answer.Picked) == 0 {
		answer.Picked = []string{q.current().Options[q.cursor].Label}
	}
	q.index, q.cursor = q.index+1, 0
	if q.index == len(q.req.Questions) {
		m.finishQuestion(q.answers)
		return
	}
	m.layout(m.windowHeight)
}

// finishQuestion hands the answers to the waiting tool, nil when dismissed,
// and writes them into the transcript.
func (m *Model) finishQuestion(answers []askuser.Answer) {
	q := m.run.question
	m.run.question = nil
	q.req.Answers <- answers
	if answers != nil {
		m.say(fromClient, askuser.Reply(q.req.Questions, answers))
	}
	herdr.Report(m.herdrClient, herdr.Working)
	m.layout(m.windowHeight)
}
