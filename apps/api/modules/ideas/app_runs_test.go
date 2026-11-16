package ideas_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/modules/ideas"
	"github.com/FacileStudio/Bulle/core/nacelletest"
	"github.com/FacileStudio/nacelle"
)

type problem struct {
	Error struct {
		Code       string `json:"code"`
		UpgradeURL string `json:"upgrade_url"`
	} `json:"error"`
}

func TestStartRunRefusesAnotherAppsName(t *testing.T) {
	a := newAppAPI(t, scripted{}, plan{paid: true})
	cases := []struct {
		name, token, body string
		want              int
	}{
		{"another app", agendaWrite, `{"app":"opus","prompt":"Plan my week"}`, http.StatusForbidden},
		{"bulle", agendaWrite, `{"app":"bulle","prompt":"Plan my week"}`, http.StatusForbidden},
		{"no app", agendaWrite, `{"prompt":"Plan my week"}`, http.StatusForbidden},
		{"a page of another app", agendaWrite, `{"app":"agenda","prompt":"Plan","page":{"app":"vision","id":"s1","title":"t","url":"https://vision.facile.studio/s/1"}}`, http.StatusBadRequest},
		{"a read token", gloutonRead, `{"app":"glouton","prompt":"Hi"}`, http.StatusForbidden},
		{"a token acting for nobody", noActor, `{"app":"glouton","prompt":"Hi"}`, http.StatusForbidden},
	}
	for _, c := range cases {
		if rec := a.call(c.token, http.MethodPost, "/api/runs", c.body); rec.Code != c.want {
			t.Errorf("%s = %d %s, want %d", c.name, rec.Code, rec.Body, c.want)
		}
	}
	if chats, _ := a.store.ListSessions(context.Background(), 1, ""); len(chats) != 0 {
		t.Fatalf("a refused run left %d chats", len(chats))
	}
}

func TestStartRunChecksThePromptAndThePage(t *testing.T) {
	a := newAppAPI(t, scripted{}, plan{paid: true})
	long := strings.Repeat("a", store.IdeaPromptLimit+1)
	cases := []string{
		`{"app":"agenda","prompt":"  "}`,
		`{"app":"agenda","prompt":"` + long + `"}`,
		`{"app":"agenda","prompt":"Plan","page":{"app":"agenda","id":"w","title":"t","url":"javascript:alert(1)"}}`,
		`{"app":"agenda","prompt":"Plan","extra":1}`,
		`{"app":"agenda","prompt":"Plan","space_id":"missing"}`,
	}
	for _, body := range cases {
		rec := a.call(agendaWrite, http.MethodPost, "/api/runs", body)
		if rec.Code != http.StatusBadRequest && rec.Code != http.StatusNotFound {
			t.Errorf("%.60s = %d %s", body, rec.Code, rec.Body)
		}
	}
}

func TestAFreePlanCannotStartARun(t *testing.T) {
	a := newAppAPI(t, scripted{}, plan{paid: false})
	rec := a.call(agendaWrite, http.MethodPost, "/api/runs", `{"app":"agenda","prompt":"Plan my week"}`)
	var body problem
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if rec.Code != http.StatusPaymentRequired || body.Error.Code != "payment_required" || body.Error.UpgradeURL != "https://registre.facile.studio/spaces" {
		t.Fatalf("free plan run = %d %s", rec.Code, rec.Body)
	}
	if chats, _ := a.store.ListSessions(context.Background(), 1, ""); len(chats) != 0 {
		t.Fatalf("a refused run left %d chats", len(chats))
	}
}

func TestAnAppRunStartsAnOwnerChatAndWaitsForApprovals(t *testing.T) {
	backend := scripted{turns: []nacelletest.Turn{
		{Calls: []nacelle.ToolCall{{ID: "c1", Name: "opus_create_task", Input: json.RawMessage(`{"title":"Call Acme"}`)}}},
		{Text: "Created the task Call Acme."},
	}}
	a := newAppAPI(t, backend, plan{paid: true})
	ctx := context.Background()
	space, _ := a.store.EnsureDefaultSpace(ctx, 1)
	body := `{"app":"Agenda","space_id":"` + space.ID + `","prompt":"Plan my week","page":{"app":"agenda","id":"week","title":"Week","url":"https://agenda.facile.studio/calendar"}}`
	rec := a.call(agendaWrite, http.MethodPost, "/api/runs", body)
	var started ideas.RunStartResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &started)
	if rec.Code != http.StatusAccepted || started.RunID == "" || started.SessionID == "" || started.IdeaID != "" || strings.Contains(rec.Body.String(), "idea_id") {
		t.Fatalf("start = %d %s", rec.Code, rec.Body)
	}
	waiting := a.pollRun(started.RunID, func(r ideas.RunResponse) bool { return r.Status == runner.RunWaiting })
	if len(waiting.PendingApprovals) != 1 || waiting.PendingApprovals[0].Tool != "opus_create_task" || waiting.SessionID != started.SessionID {
		t.Fatalf("waiting run = %+v", waiting)
	}
	if rec := a.call(gloutonWrite, http.MethodPost, waiting.PendingApprovals[0].DecidePath, `{"approve":true}`); rec.Code != http.StatusOK {
		t.Fatalf("approve = %d %s", rec.Code, rec.Body)
	}
	done := a.pollRun(started.RunID, func(r ideas.RunResponse) bool { return r.Status == store.RunFinished })
	if done.Summary != "Created the task Call Acme." || len(done.Outputs) != 1 {
		t.Fatalf("finished run = %+v", done)
	}
	chat, err := a.store.GetSession(ctx, 1, started.SessionID)
	if err != nil || chat.Channel != store.ChannelWeb || chat.SpaceID != space.ID || chat.Title != "Plan my week" {
		t.Fatalf("chat = %+v %v", chat, err)
	}
}

func TestTheAssistantProfileIsFreeAndMinimal(t *testing.T) {
	a := newAppAPI(t, scripted{}, plan{paid: false})
	rec := a.call(gloutonRead, http.MethodGet, "/api/assistant/profile", "")
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != `{"name":"","avatar_url":""}` {
		t.Fatalf("empty profile = %d %s", rec.Code, rec.Body)
	}
	if err := a.store.SaveAssistantProfile(context.Background(), store.AssistantProfile{UserID: 1, Name: "Nova", AvatarURL: "https://x.test/a.png", Archetype: "calm", Traits: []string{"dry"}}); err != nil {
		t.Fatal(err)
	}
	rec = a.call(gloutonRead, http.MethodGet, "/api/assistant/profile", "")
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != `{"name":"Nova","avatar_url":"https://x.test/a.png"}` {
		t.Fatalf("profile = %d %s", rec.Code, rec.Body)
	}
	for name, token := range map[string]string{"another scope": otherScope, "no actor": noActor} {
		if rec := a.call(token, http.MethodGet, "/api/assistant/profile", ""); rec.Code != http.StatusForbidden {
			t.Errorf("%s = %d", name, rec.Code)
		}
	}
}
