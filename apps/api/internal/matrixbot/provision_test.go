package matrixbot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/prisme"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

const testIssuer = "https://sso.test"

// fakePrisme records what Bulle sends: the rooms joined, the agent's name and
// the X-Prisme-Acting-Subject header of every call, "" when absent. With
// refuse set, POST /api/agents answers 409 with that error code.
type fakePrisme struct {
	joined    []string
	agentName string
	subjects  []string
	refuse    string
	agents    int
}

func servePrisme(t *testing.T, f *fakePrisme) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.subjects = append(f.subjects, r.Header.Get("X-Prisme-Acting-Subject"))
		switch r.Method + " " + r.URL.EscapedPath() {
		case "POST /api/agents":
			f.agents++
			if f.refuse != "" {
				w.WriteHeader(http.StatusConflict)
				_, _ = fmt.Fprintf(w, `{"error":{"code":%q,"message":"the subject belongs to another account"}}`, f.refuse)
				return
			}
			var in map[string]string
			_ = json.NewDecoder(r.Body).Decode(&in)
			f.agentName = in["name"]
			_, _ = w.Write([]byte(`{"homeserver":"http://127.0.0.1:1","user_id":"@agent-me-pixel:p","access_token":"tok","device_id":"D"}`))
		case "GET /api/spaces":
			_, _ = w.Write([]byte(`[{"id":1,"name":"Work","channels":[{"room_id":"!w:p","name":"general"}]},` +
				`{"id":2,"name":"Side","channels":[{"room_id":"!s:p","name":"general"}]}]`))
		case "PUT /api/agents/@agent-me-pixel:p/rooms":
			var in map[string][]string
			_ = json.NewDecoder(r.Body).Decode(&in)
			f.joined = in["room_ids"]
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func seedProvision(t *testing.T, subject string) {
	t.Helper()
	queries := []string{
		"TRUNCATE users RESTART IDENTITY CASCADE",
		"INSERT INTO users (id, email) VALUES (1, 'me@facile.studio')",
		"INSERT INTO spaces (id, user_id, name) VALUES ('home', 1, 'Personal'), ('work', 1, 'Work')",
		"INSERT INTO assistant_profiles (user_id, name) VALUES (1, 'Pixel')",
	}
	if subject != "" {
		queries = append(queries, "INSERT INTO porte_identities (user_id, provider, subject) VALUES (1, '"+testIssuer+"', '"+subject+"')")
	}
	for _, q := range queries {
		if err := testDB.Exec(q).Error; err != nil {
			t.Fatal(err)
		}
	}
}

func provisionManager(t *testing.T, srv *httptest.Server, logs io.Writer) *Manager {
	t.Helper()
	if testDB == nil {
		t.Skip("no TEST_DATABASE_URL")
	}
	sqlDB, err := testDB.DB()
	if err != nil {
		t.Fatal(err)
	}
	m, err := New(Config{
		SQL: sqlDB, Store: store.NewPostgres(testDB), Key: bytes.Repeat([]byte{1}, 32), Issuer: testIssuer,
		Prisme: prisme.New(srv.URL, "key"), Logger: slog.New(slog.NewTextHandler(logs, nil)),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(m.Stop)
	return m
}

func TestProvisionCreatesTheAgentAndMapsRoomsBySpaceName(t *testing.T) {
	f := &fakePrisme{}
	m := provisionManager(t, servePrisme(t, f), io.Discard)
	seedProvision(t, "sub-1")
	ctx := context.Background()
	m.provisionAll(ctx)
	if f.agentName != "Pixel" || len(f.joined) != 2 {
		t.Fatalf("agent %q joined %v, want Pixel in both channels", f.agentName, f.joined)
	}
	if len(f.subjects) != 3 || strings.Join(f.subjects, ",") != "sub-1,sub-1,sub-1" {
		t.Fatalf("every call must carry the user's Registre subject, got %q", f.subjects)
	}
	if space, _ := m.repo.roomSpace(ctx, 1, "!w:p"); space != "work" {
		t.Errorf("!w:p works in %q, want the Work space", space)
	}
	if space, _ := m.repo.roomSpace(ctx, 1, "!s:p"); space != "home" {
		t.Errorf("!s:p works in %q, want the default space", space)
	}
	status, err := m.Get(ctx, 1)
	if err != nil || status.User != "@agent-me-pixel:p" || len(status.Rooms) != 2 || !status.Available {
		t.Fatalf("status = %+v, %v", status, err)
	}
	bot, _, _ := m.repo.get(ctx, 1)
	if token, err := m.seal.open(1, bot.Token, bot.Nonce); err != nil || token != "tok" {
		t.Fatalf("stored token = %q, %v", token, err)
	}
}

func TestProvisionSendsNoSubjectWithoutAnIdentity(t *testing.T) {
	f := &fakePrisme{}
	m := provisionManager(t, servePrisme(t, f), io.Discard)
	seedProvision(t, "")
	if err := testDB.Exec("INSERT INTO porte_identities (user_id, provider, subject) VALUES (1, 'https://other.test', 'elsewhere')").Error; err != nil {
		t.Fatal(err)
	}
	m.provisionAll(context.Background())
	if f.agentName != "Pixel" || len(f.subjects) != 3 || strings.Join(f.subjects, "") != "" {
		t.Fatalf("an identity for another issuer is never guessed, got subjects %q", f.subjects)
	}
}

func TestProvisionHoldsAMismatchedSubjectUntilRestart(t *testing.T) {
	f := &fakePrisme{refuse: "acting_subject_mismatch"}
	var logs bytes.Buffer
	m := provisionManager(t, servePrisme(t, f), &logs)
	seedProvision(t, "sub-1")
	ctx := context.Background()
	clock := time.Date(2026, 11, 11, 8, 0, 0, 0, time.UTC)
	m.now = func() time.Time { return clock }
	m.provisionAll(ctx)
	m.provisionAll(ctx)
	clock = clock.Add(72 * time.Hour)
	m.provisionAll(ctx)
	if f.agents != 1 {
		t.Fatalf("a mismatched subject must never be retried, Prisme saw %d agent calls", f.agents)
	}
	lines := strings.Count(logs.String(), "code=acting_subject_mismatch")
	if lines != 1 || !strings.Contains(logs.String(), "user_id=1") || !strings.Contains(logs.String(), "level=ERROR") || strings.Contains(logs.String(), "level=WARN") {
		t.Fatalf("the refusal must be logged once at ERROR with the user id and the code, got %d lines:\n%s", lines, logs.String())
	}
	if !m.onHold(1) {
		t.Fatal("the user is not on hold")
	}
}

func TestProvisionRetriesATakenSubjectDaily(t *testing.T) {
	f := &fakePrisme{refuse: prisme.CodeSubjectTaken}
	var logs bytes.Buffer
	m := provisionManager(t, servePrisme(t, f), &logs)
	seedProvision(t, "sub-1")
	ctx := context.Background()
	clock := time.Date(2026, 11, 11, 8, 0, 0, 0, time.UTC)
	m.now = func() time.Time { return clock }
	m.provisionAll(ctx)
	m.provisionAll(ctx)
	clock = clock.Add(23 * time.Hour)
	m.provisionAll(ctx)
	if f.agents != 1 || !m.onHold(1) {
		t.Fatalf("a taken subject is left alone for a day, Prisme saw %d agent calls", f.agents)
	}
	clock = clock.Add(2 * time.Hour)
	m.provisionAll(ctx)
	if f.agents != 2 || strings.Count(logs.String(), "code=acting_subject_taken") != 2 || strings.Contains(logs.String(), "level=ERROR") {
		t.Fatalf("after a day the user is retried and the refusal logged again at WARN, got %d calls:\n%s", f.agents, logs.String())
	}
	f.refuse = ""
	clock = clock.Add(25 * time.Hour)
	m.provisionAll(ctx)
	if f.agents != 3 || f.agentName != "Pixel" || m.onHold(1) {
		t.Fatalf("once Prisme accepts the subject the user is provisioned and no longer held, got %d calls, agent %q", f.agents, f.agentName)
	}
}

func TestAgentAvatarKeepsOnlyHTTPURLs(t *testing.T) {
	for raw, want := range map[string]string{
		"":                                   "",
		"  ":                                 "",
		"https://cdn.facile.studio/a.png":    "https://cdn.facile.studio/a.png",
		" http://example.com/me.jpg ":        "http://example.com/me.jpg",
		"data:image/svg+xml;base64,PHN2Zz4=": "",
		"javascript:alert(1)":                "",
		"mxc://matrix.org/abc":               "",
		"https://":                           "",
	} {
		if got := agentAvatar(raw); got != want {
			t.Errorf("agentAvatar(%q) = %q, want %q", raw, got, want)
		}
	}
}
