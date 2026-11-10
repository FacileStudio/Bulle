package prisme

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientActsForTheUserWithTheServiceKey(t *testing.T) {
	var rooms, subjects []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer key" || r.Header.Get("X-Prisme-Acting-For") != "me@facile.studio" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		subjects = append(subjects, r.Header.Get("X-Prisme-Acting-Subject"))
		switch r.Method + " " + r.URL.EscapedPath() {
		case "POST /api/agents":
			if r.Header.Get("X-Prisme-Acting-Subject") == "someone-else" {
				w.WriteHeader(http.StatusConflict)
				_, _ = w.Write([]byte(`{"error":{"code":"identity_conflict","message":"the subject belongs to another account"}}`))
				return
			}
			_, _ = w.Write([]byte(`{"homeserver":"https://p","user_id":"@agent-me-pixel:p","access_token":"t","device_id":"D"}`))
		case "GET /api/spaces":
			_, _ = w.Write([]byte(`[{"id":1,"room_id":"!space","name":"Work","slug":"work","channels":[{"room_id":"!a","name":"general","alias":"#prisme-work-general:p"}]}]`))
		case "PUT /api/agents/@agent-me-pixel:p/rooms":
			var in struct {
				RoomIDs []string `json:"room_ids"`
			}
			_ = json.NewDecoder(r.Body).Decode(&in)
			rooms = in.RoomIDs
		case "PUT /api/agents/broken/rooms":
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("nope"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	c := New(srv.URL+"/", "key")
	ctx := context.Background()
	me := Actor{Email: "me@facile.studio", Subject: "sub-1"}
	acc, err := c.Agent(ctx, me, "Pixel", "")
	if err != nil || acc.AccessToken != "t" || acc.DeviceID != "D" {
		t.Fatalf("agent = %+v, %v", acc, err)
	}
	spaces, err := c.Spaces(ctx, Actor{Email: "me@facile.studio"})
	if err != nil || len(spaces) != 1 || spaces[0].Channels[0].RoomID != "!a" {
		t.Fatalf("spaces = %+v, %v", spaces, err)
	}
	if err := c.SetRooms(ctx, me, acc.UserID, []string{"!a:p"}); err != nil || len(rooms) != 1 {
		t.Fatalf("set rooms = %v %v", rooms, err)
	}
	if want := []string{"sub-1", "", "sub-1"}; len(subjects) != 3 || subjects[0] != want[0] || subjects[1] != want[1] || subjects[2] != want[2] {
		t.Fatalf("subject headers = %q, want %q: sent only when the actor has one", subjects, want)
	}
	if _, err := New(srv.URL, "wrong").Spaces(ctx, me); err == nil {
		t.Fatal("a wrong key was accepted")
	}
	if New("", "key") != nil || New(srv.URL, "") != nil {
		t.Fatal("an unconfigured client is not nil")
	}
}

func TestClientReadsPrismeRefusals(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Header.Get("X-Prisme-Acting-Subject") == "someone-else":
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"error":{"code":"identity_conflict","message":"the subject belongs to another account"}}`))
		case r.Header.Get("X-Prisme-Acting-Subject") == "bad":
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"code":"invalid_argument","message":"malformed subject"}}`))
		default:
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("nope"))
		}
	}))
	defer srv.Close()
	c := New(srv.URL, "key")
	ctx := context.Background()
	var refusal *Refusal
	_, err := c.Agent(ctx, Actor{Email: "me@facile.studio", Subject: "someone-else"}, "Pixel", "")
	if !errors.As(err, &refusal) || refusal.Status != 409 || refusal.Code != "identity_conflict" || !refusal.Final() {
		t.Fatalf("a 409 JSON error must be a final refusal, got %v", err)
	}
	if refusal.Error() != "prisme POST /api/agents: 409 identity_conflict: the subject belongs to another account" {
		t.Fatalf("refusal message = %q", refusal.Error())
	}
	_, err = c.Spaces(ctx, Actor{Email: "me@facile.studio", Subject: "bad"})
	if !errors.As(err, &refusal) || refusal.Status != 400 || !refusal.Final() {
		t.Fatalf("a 400 JSON error must be a final refusal, got %v", err)
	}
	err = c.SetRooms(ctx, Actor{Email: "me@facile.studio"}, "agent", nil)
	if err == nil || errors.As(err, &refusal) {
		t.Fatalf("a plain 500 keeps today's error, got %v", err)
	}
}
