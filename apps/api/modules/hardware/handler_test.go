package hardware_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	internal "github.com/FacileStudio/Bulle/apps/api/internal/hardware"
	"github.com/FacileStudio/Bulle/apps/api/modules/hardware"
	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
)

func server(t *testing.T, st *fakeStore) *httptest.Server {
	t.Helper()
	asUser := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var id int64 = 1
			if r.Header.Get("X-User") == "2" {
				id = 2
			}
			next.ServeHTTP(w, r.WithContext(authcontext.WithIdentity(r.Context(), authcontext.Identity{UserID: id})))
		})
	}
	router := chi.NewRouter()
	hardware.RegisterRoutes(router, hardware.Config{Devices: st, Sockets: st, Registry: internal.NewRegistry()}, asUser)
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	return srv
}

func send(t *testing.T, method, url, user, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User", user)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = res.Body.Close() })
	return res
}

func claim(t *testing.T, srv *httptest.Server) hardware.ClaimResponse {
	t.Helper()
	res := send(t, http.MethodPost, srv.URL+"/hardware/devices/claim", "1", `{"hardware_id":"4f2a91c0"}`)
	var claimed hardware.ClaimResponse
	if err := json.NewDecoder(res.Body).Decode(&claimed); err != nil || res.StatusCode != http.StatusCreated {
		t.Fatalf("claim = %d, %v", res.StatusCode, err)
	}
	return claimed
}

func TestClaimReturnsTheTokenOnceAndKeepsItsHash(t *testing.T) {
	st := newFakeStore()
	srv := server(t, st)
	claimed := claim(t, srv)
	if !internal.IsToken(claimed.DeviceToken) || !strings.HasPrefix(claimed.Device.ID, internal.IDPrefix) {
		t.Fatalf("claim answered %+v", claimed)
	}
	if st.hashes[internal.HashToken(claimed.DeviceToken)] != claimed.Device.ID {
		t.Fatal("the store does not hold the SHA-256 of the returned token")
	}
	if claimed.Device.HardwareID != "4F2A91C0" || claimed.Device.Name != "Pastille #91C0" || claimed.Device.DeviceClass != "charm" {
		t.Fatalf("claimed device = %+v", claimed.Device)
	}
	res := send(t, http.MethodGet, srv.URL+"/hardware/devices", "1", "")
	body := new(strings.Builder)
	if _, err := io.Copy(body, res.Body); err != nil || strings.Contains(body.String(), claimed.DeviceToken) {
		t.Fatalf("the device list leaks the token or failed: %v", err)
	}
}

func TestRevokeIsLimitedToTheOwner(t *testing.T) {
	srv := server(t, newFakeStore())
	claimed := claim(t, srv)
	url := srv.URL + "/hardware/devices/" + claimed.Device.ID
	if res := send(t, http.MethodDelete, url, "2", ""); res.StatusCode != http.StatusNotFound {
		t.Fatalf("another user's revoke = %d, want 404", res.StatusCode)
	}
	if res := send(t, http.MethodDelete, url, "1", ""); res.StatusCode != http.StatusNoContent {
		t.Fatalf("owner revoke = %d, want 204", res.StatusCode)
	}
}

func TestRelayRefusesADeviceOfAnotherUser(t *testing.T) {
	srv := server(t, newFakeStore())
	claimed := claim(t, srv)
	_, res, err := websocket.Dial(context.Background(), wsURL(srv)+"?device_id="+claimed.Device.ID, &websocket.DialOptions{
		Subprotocols: []string{"bohs.v1"},
		HTTPHeader:   http.Header{"X-User": {"2"}},
	})
	if err == nil || res == nil || res.StatusCode != http.StatusNotFound {
		t.Fatalf("relay of another user's device: %v", err)
	}
}

func TestDirectSocketAsksForToolsAndClosesOnRevoke(t *testing.T) {
	srv := server(t, newFakeStore())
	claimed := claim(t, srv)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, wsURL(srv), &websocket.DialOptions{
		Subprotocols: []string{"bohs.v1"},
		HTTPHeader:   http.Header{"Authorization": {"Bearer " + claimed.DeviceToken}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()
	kind, frame, err := conn.Read(ctx)
	if err != nil || kind != websocket.MessageBinary || frame[0] != 0x02 || !strings.Contains(string(frame[1:]), `"method":"tools/list"`) {
		t.Fatalf("first frame = %v %q, %v", kind, frame, err)
	}
	send(t, http.MethodDelete, srv.URL+"/hardware/devices/"+claimed.Device.ID, "1", "")
	_, _, err = conn.Read(ctx)
	if websocket.CloseStatus(err) != internal.StatusRevoked {
		t.Fatalf("read after revoke = %v, want close %d", err, internal.StatusRevoked)
	}
}

func wsURL(srv *httptest.Server) string {
	return "ws" + strings.TrimPrefix(srv.URL, "http") + "/hardware/ws"
}
