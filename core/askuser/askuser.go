// Package askuser holds the ask_user tool's contract, shared by the web
// runner and the terminal: the questions a model may ask, how they are
// checked, and how the user's answers read back.
package askuser

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Name is the tool's name on every surface.
const Name = "ask_user"

// Description is what every surface tells the model; each adds how the
// answers come back.
const Description = "Ask the user one to four multiple-choice questions, when you cannot go on without a choice only they can make: " +
	"an ambiguous request, two valid approaches with a real trade-off, or a missing detail you cannot find with your tools. " +
	"Do not use it for a simple yes or no, for something you can look up, or to confirm an action that already asks for approval. " +
	"Give each question two to four distinct options, the one you recommend first. The user can always type their own answer instead, so never add an Other option."

// Option is one choice of a question.
type Option struct {
	Label       string `json:"label" jsonschema:"required,description=The choice in one to five words"`
	Description string `json:"description,omitempty" jsonschema:"description=What picking it means or costs, in one short sentence"`
}

// Question is one question with its choices.
type Question struct {
	Question    string   `json:"question" jsonschema:"required,description=The full question, ending with a question mark"`
	Header      string   `json:"header,omitempty" jsonschema:"description=A label of at most 12 characters, such as Date or Tone"`
	Options     []Option `json:"options" jsonschema:"required,description=Two to four distinct choices, the recommended one first"`
	MultiSelect bool     `json:"multi_select,omitempty" jsonschema:"description=True when the user may pick several options"`
}

// Input is the tool's argument.
type Input struct {
	Questions []Question `json:"questions" jsonschema:"required,description=One to four questions"`
}

// Answer is what the user gave one question: the options they picked and
// what they typed, either of which may be empty.
type Answer struct {
	Picked []string
	Typed  string
}

// Validate refuses a card the user could not answer well: no question or
// more than four, an empty question, fewer than two or more than four
// options, an empty or repeated option, or an Other option of the model's.
func (in Input) Validate() error {
	if len(in.Questions) == 0 || len(in.Questions) > 4 {
		return errors.New("ask one to four questions")
	}
	for i, q := range in.Questions {
		if strings.TrimSpace(q.Question) == "" {
			return fmt.Errorf("question %d is empty", i+1)
		}
		if err := checkOptions(q.Options); err != nil {
			return fmt.Errorf("question %d: %w", i+1, err)
		}
	}
	return nil
}

func checkOptions(options []Option) error {
	if len(options) < 2 || len(options) > 4 {
		return errors.New("give two to four options")
	}
	seen := map[string]bool{}
	for _, option := range options {
		label := strings.ToLower(strings.TrimSpace(option.Label))
		switch {
		case label == "" || seen[label]:
			return errors.New("an option is empty or repeated")
		case label == "other":
			return errors.New("drop the Other option, the user can always type their own answer")
		}
		seen[label] = true
	}
	return nil
}

// Reply is the answers as the model reads them, one line per question. A
// typed answer wins over a pick for a single-choice question and joins the
// picks of a multi-choice one.
func Reply(questions []Question, answers []Answer) string {
	lines := make([]string, len(questions))
	for i, q := range questions {
		var answer Answer
		if i < len(answers) {
			answer = answers[i]
		}
		lines[i] = q.Question + " " + cmp.Or(value(q, answer), "(no answer)")
	}
	return strings.Join(lines, "\n")
}

func value(q Question, answer Answer) string {
	typed := strings.TrimSpace(answer.Typed)
	if !q.MultiSelect {
		if typed != "" || len(answer.Picked) == 0 {
			return typed
		}
		return answer.Picked[0]
	}
	parts := append([]string{}, answer.Picked...)
	if typed != "" {
		parts = append(parts, typed)
	}
	return strings.Join(parts, ", ")
}

// UnmarshalJSON accepts multiSelect, the spelling Claude Code's tool uses, as
// an alias of multi_select: either one set to true allows several picks.
func (q *Question) UnmarshalJSON(data []byte) error {
	type plain Question
	var alias struct {
		MultiSelect bool `json:"multiSelect"`
	}
	if err := json.Unmarshal(data, (*plain)(q)); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	q.MultiSelect = q.MultiSelect || alias.MultiSelect
	return nil
}
