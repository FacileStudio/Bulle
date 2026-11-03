package matrixbot

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/prisme"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

func fakePrisme(t *testing.T, joined *[]string, agentName *string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method + " " + r.URL.EscapedPath() {
		case "POST /api/agents":
			var in map[string]string
			_ = json.NewDecoder(r.Body).Decode(&in)
			*agentName = in["name"]
			_, _ = w.Write([]byte(`{"homeserver":"http://127.0.0.1:1","user_id":"@agent-me-pixel:p","access_token":"tok","device_id":"D"}`))
		case "GET /api/spaces":
			_, _ = w.Write([]byte(`[{"id":"a","name":"Work","channels":[{"room_id":"!w:p","name":"general"}]},` +
				`{"id":"b","name":"Side","channels":[{"room_id":"!s:p","name":"general"}]}]`))
		case "PUT /api/agents/@agent-me-pixel:p/rooms":
			var in map[string][]string
			_ = json.NewDecoder(r.Body).Decode(&in)
			*joined = in["room_ids"]
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func seedProvision(t *testing.T) {
	t.Helper()
	for _, q := range []string{
		"TRUNCATE users RESTART IDENTITY CASCADE",
		"INSERT INTO users (id, email) VALUES (1, 'me@facile.studio')",
		"INSERT INTO spaces (id, user_id, name) VALUES ('home', 1, 'Personal'), ('work', 1, 'Work')",
		"INSERT INTO assistant_profiles (user_id, name) VALUES (1, 'Pixel')",
	} {
		if err := testDB.Exec(q).Error; err != nil {
			t.Fatal(err)
		}
	}
}

func TestProvisionCreatesTheAgentAndMapsRoomsBySpaceName(t *testing.T) {
	if testDB == nil {
		t.Skip("no TEST_DATABASE_URL")
	}
	seedProvision(t)
	var joined []string
	var agentName string
	srv := fakePrisme(t, &joined, &agentName)
	sqlDB, err := testDB.DB()
	if err != nil {
		t.Fatal(err)
	}
	m, err := New(Config{
		SQL: sqlDB, Store: store.NewPostgres(testDB), Key: bytes.Repeat([]byte{1}, 32),
		Prisme: prisme.New(srv.URL, "key"), Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Stop()
	ctx := context.Background()
	if err := m.provision(ctx, user{ID: 1, Email: "me@facile.studio"}); err != nil {
		t.Fatal(err)
	}
	if agentName != "Pixel" || len(joined) != 2 {
		t.Fatalf("agent %q joined %v, want Pixel in both channels", agentName, joined)
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
