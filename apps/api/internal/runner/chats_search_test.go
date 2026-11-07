package runner_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

func TestChatsSearchSeesOnlyThisSpace(t *testing.T) {
	backend := &fakeBackend{calls: []call{{id: "s1", name: runner.ChatsSearchName, input: `{"query":"dentist"}`}}}
	f := newFixture(t, backend, nil, nil)
	ctx := context.Background()
	work, err := f.store.CreateSpace(ctx, store.Space{UserID: 1, Name: "Work"})
	if err != nil {
		t.Fatal(err)
	}
	home := chatWith(t, f, f.session.SpaceID, "Book the dentist on Friday")
	elsewhere := chatWith(t, f, work.ID, "The dentist invoice is paid")
	runID, err := f.runner.Start(ctx, 1, f.session.ID, "when is my dentist?")
	if err != nil {
		t.Fatal(err)
	}
	f.waitDone(t, runID)
	events, _ := f.store.Events(ctx, runID, 0)
	var output string
	for _, event := range events {
		var end map[string]any
		if event.Type == "TOOL_CALL_END" && json.Unmarshal(event.Payload, &end) == nil && end["call_id"] == "s1" {
			output, _ = end["output_preview"].(string)
		}
	}
	if !strings.Contains(output, home.ID) || strings.Contains(output, elsewhere.ID) || !strings.Contains(output, "Book the dentist") {
		t.Fatalf("chats_search answered outside its space or missed the chat: %q", output)
	}
}

func chatWith(t *testing.T, f *fixture, spaceID string, texts ...string) store.Session {
	t.Helper()
	ctx := context.Background()
	session, err := f.store.CreateSession(ctx, store.Session{ID: store.NewID(), UserID: 1, SpaceID: spaceID, Channel: store.ChannelWeb, Title: "Chat"})
	if err != nil {
		t.Fatal(err)
	}
	payloads := make([]json.RawMessage, len(texts))
	for i, text := range texts {
		role := nacelle.RoleUser
		if i%2 == 1 {
			role = nacelle.RoleAssistant
		}
		payloads[i], err = runner.EncodeMessage(nacelle.Message{Role: role, Parts: []nacelle.Part{nacelle.Text{Text: text}}}, nil)
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := f.store.ReplaceMessages(ctx, session.ID, payloads); err != nil {
		t.Fatal(err)
	}
	return session
}
