package ask_test

import (
	"context"
	"encoding/json"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/FacileStudio/Bulle/apps/cli/internal/ask"
	"github.com/FacileStudio/Bulle/core/askuser"
)

const card = `{"questions":[{"question":"Which day?","options":[{"label":"Monday"},{"label":"Tuesday"}]}]}`

func TestToolWaitsForTheAnswersAndReturnsThem(t *testing.T) {
	asker := ask.New()
	asker.Wire(func(msg tea.Msg) {
		req := msg.(ask.Request)
		req.Answers <- []askuser.Answer{{Picked: []string{req.Questions[0].Options[1].Label}}}
	})
	tool, err := asker.Tool()
	if err != nil {
		t.Fatal(err)
	}
	out, err := tool.Run(context.Background(), json.RawMessage(card))
	if err != nil || out != "Which day? Tuesday" {
		t.Fatalf("out %q err %v", out, err)
	}
}

func TestToolRefusesWithNobodyToAnswer(t *testing.T) {
	tool, err := ask.New().Tool()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tool.Run(context.Background(), json.RawMessage(card)); err == nil {
		t.Fatal("an unwired asker answered")
	}
}
