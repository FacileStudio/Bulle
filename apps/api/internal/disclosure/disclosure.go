// Package disclosure decides whether an agent's reply may be posted where
// people other than its owner can read it. It asks a decision model whether the
// text reveals private data and fails closed: no key, a failed call, or an
// unsure verdict all hold the reply back.
//
// The screen is the second wall, never the first. What an agent may read is
// settled before the run by who asked and where (see matrixbot's clearance); this
// only catches what a permitted run says in a room that is wider than the owner.
package disclosure

import (
	"context"

	"github.com/FacileStudio/Bulle/core/jev"
)

const (
	question   = "disclosure"
	optionSafe = "shareable"
	optionLeak = "private"
	floor      = 0.6
	maxText    = 16 << 10
)

// Screen vets replies with one decision call each.
type Screen struct {
	client *jev.Client
}

// New builds a screen on the given jev settings. A Config with no key falls back
// to TYPESAFE_API_KEY; with neither, every reply is held back.
func New(cfg jev.Config) *Screen {
	return &Screen{client: jev.New(cfg)}
}

// Allows reports whether text may be shown to people who are not the agent's
// owner. Only a confident "shareable" verdict lets it through.
func (s *Screen) Allows(ctx context.Context, text string) bool {
	if s == nil || s.client == nil {
		return false
	}
	response, err := s.client.Evaluate(ctx, clamp(text), map[string]jev.Question{question: ask()})
	if err != nil {
		return false
	}
	answer := response.Answers[question]
	return answer.Choice == optionSafe && answer.Confidence >= floor
}

func ask() jev.Question {
	return jev.Question{
		Type: "choice",
		Instructions: "The state is a message an AI assistant is about to post in a team chat room, where colleagues of the person it works for will read it. " +
			"Decide whether the message may be shown to those colleagues.",
		Criteria: map[string]string{
			optionLeak: "it reveals private or confidential data: the content of emails or messages, calendar entries, contacts, phone numbers or home addresses, " +
				"passwords, tokens or keys, money, health or HR matters, private files or notes, or anything else only the assistant's owner should see",
			optionSafe: "it holds nothing private: general conversation, public knowledge, or work information any colleague in the room may read",
		},
		Options: []string{optionSafe, optionLeak},
	}
}

func clamp(text string) string {
	if len(text) <= maxText {
		return text
	}
	return text[:maxText]
}
