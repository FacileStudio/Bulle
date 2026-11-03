package prisme

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientActsForTheUserWithTheServiceKey(t *testing.T) {
	var rooms []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer key" || r.Header.Get("X-Prisme-Acting-For") != "me@facile.studio" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch r.Method + " " + r.URL.EscapedPath() {
		case "POST /api/agents":
			_, _ = w.Write([]byte(`{"homeserver":"https://p","user_id":"@agent-me-pixel:p","access_token":"t","device_id":"D"}`))
		case "GET /api/spaces":
			_, _ = w.Write([]byte(`[{"id":1,"room_id":"!space","name":"Work","slug":"work","channels":[{"room_id":"!a","name":"general","alias":"#prisme-work-general:p"}]}]`))
		case "PUT /api/agents/@agent-me-pixel:p/rooms":
			var in struct {
				RoomIDs []string `json:"room_ids"`
			}
			_ = json.NewDecoder(r.Body).Decode(&in)
			rooms = in.RoomIDs
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	c := New(srv.URL+"/", "key")
	ctx := context.Background()
	acc, err := c.Agent(ctx, "me@facile.studio", "Pixel", "")
	if err != nil || acc.AccessToken != "t" || acc.DeviceID != "D" {
		t.Fatalf("agent = %+v, %v", acc, err)
	}
	spaces, err := c.Spaces(ctx, "me@facile.studio")
	if err != nil || len(spaces) != 1 || spaces[0].Channels[0].RoomID != "!a" {
		t.Fatalf("spaces = %+v, %v", spaces, err)
	}
	if err := c.SetRooms(ctx, "me@facile.studio", acc.UserID, []string{"!a:p"}); err != nil || len(rooms) != 1 {
		t.Fatalf("set rooms = %v %v", rooms, err)
	}
	if _, err := New(srv.URL, "wrong").Spaces(ctx, "me@facile.studio"); err == nil {
		t.Fatal("a wrong key was accepted")
	}
	if New("", "key") != nil || New(srv.URL, "") != nil {
		t.Fatal("an unconfigured client is not nil")
	}
}
