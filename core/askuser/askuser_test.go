package askuser_test

import (
	"testing"

	"github.com/FacileStudio/Bulle/core/askuser"
)

func day() askuser.Question {
	return askuser.Question{Question: "Which day?", Options: []askuser.Option{{Label: "Monday"}, {Label: "Tuesday"}}}
}

func TestValidateRefusesCardsTheUserCannotAnswerWell(t *testing.T) {
	bad := []askuser.Input{
		{},
		{Questions: []askuser.Question{{Question: "Which day?", Options: []askuser.Option{{Label: "Monday"}}}}},
		{Questions: []askuser.Question{{Question: "Which day?", Options: []askuser.Option{{Label: "Monday"}, {Label: "Other"}}}}},
		{Questions: []askuser.Question{{Question: "Which day?", Options: []askuser.Option{{Label: "Monday"}, {Label: "monday"}}}}},
	}
	for _, in := range bad {
		if in.Validate() == nil {
			t.Errorf("accepted %+v", in)
		}
	}
	if err := (askuser.Input{Questions: []askuser.Question{day()}}).Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestReplyPutsTypedTextFirstForOneChoice(t *testing.T) {
	tone := askuser.Question{Question: "Tone?", MultiSelect: true, Options: []askuser.Option{{Label: "Warm"}, {Label: "Dry"}}}
	got := askuser.Reply([]askuser.Question{day(), tone, day()}, []askuser.Answer{
		{Picked: []string{"Monday"}, Typed: "Friday at 3"},
		{Picked: []string{"Warm", "Dry"}, Typed: "short"},
	})
	want := "Which day? Friday at 3\nTone? Warm, Dry, short\nWhich day? (no answer)"
	if got != want {
		t.Fatalf("got %q", got)
	}
}
