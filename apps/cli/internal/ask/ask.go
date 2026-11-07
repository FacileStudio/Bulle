// Package ask gives the terminal agent the ask_user tool: the model's
// multiple-choice questions go to the running program, and the tool waits
// for the user's answers and returns them as its result.
package ask

import (
	"context"
	"errors"
	"sync"

	tea "charm.land/bubbletea/v2"

	"github.com/FacileStudio/Bulle/core/askuser"
	"github.com/FacileStudio/nacelle"
)

const description = askuser.Description + " The terminal shows the questions and waits: the tool returns the user's answers, so go on from them."

const dismissed = "The user dismissed the questions without answering. Go on with your best judgment, or ask in plain text."

// Request is a set of questions waiting for the user. The program answers
// once on Answers: one Answer per question, or nil when the user dismissed
// them. Done closes once the tool stops waiting.
type Request struct {
	Questions []askuser.Question
	Answers   chan<- []askuser.Answer
	Done      <-chan struct{}
}

// Asker hands ask_user calls to the running program, one set at a time.
type Asker struct {
	asking sync.Mutex
	send   func(tea.Msg)
}

// New creates an Asker with no program yet.
func New() *Asker { return &Asker{} }

// Wire connects the Asker to the running program's message delivery.
func (a *Asker) Wire(send func(tea.Msg)) {
	if a != nil {
		a.send = send
	}
}

// Tool is the ask_user tool backed by this Asker.
func (a *Asker) Tool() (nacelle.Tool, error) {
	return nacelle.NewToolWithOptions(askuser.Name, description, a.ask, nacelle.ToolOptions{ReadOnly: true})
}

func (a *Asker) ask(ctx context.Context, in askuser.Input) (string, error) {
	if err := in.Validate(); err != nil {
		return "", err
	}
	if a.send == nil {
		return "", errors.New("nobody can answer questions in this session; ask in plain text")
	}
	a.asking.Lock()
	defer a.asking.Unlock()
	answers := make(chan []askuser.Answer, 1)
	a.send(Request{Questions: in.Questions, Answers: answers, Done: ctx.Done()})
	select {
	case got := <-answers:
		if got == nil {
			return dismissed, nil
		}
		return askuser.Reply(in.Questions, got), nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
