package push

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type memDevices struct{ devices []Device }

func (m *memDevices) SaveDevice(_ context.Context, d Device) error {
	m.devices = append(m.devices, d)
	return nil
}

func (m *memDevices) DeleteDevice(_ context.Context, userID int64, token string) error {
	kept := m.devices[:0]
	for _, d := range m.devices {
		if d.UserID != userID || d.Token != token {
			kept = append(kept, d)
		}
	}
	m.devices = kept
	return nil
}

func (m *memDevices) Devices(_ context.Context, userID int64) ([]Device, error) {
	var out []Device
	for _, d := range m.devices {
		if d.UserID == userID {
			out = append(out, d)
		}
	}
	return out, nil
}

func TestHubDeliversToFCMAndForgetsGoneTokens(t *testing.T) {
	var got []string
	fcm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		got = append(got, string(body))
		if strings.Contains(string(body), `"token":"dead"`) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"status":"NOT_FOUND","details":[{"errorCode":"UNREGISTERED"}]}}`))
		}
	}))
	defer fcm.Close()

	devices := &memDevices{}
	hub := NewHub(HubConfig{FCM: &FCM{endpoint: fcm.URL, client: fcm.Client()}, Devices: devices})
	ctx := context.Background()
	for _, token := range []string{"live", "dead"} {
		if err := hub.Register(ctx, Device{UserID: 1, Kind: KindFCM, Token: token}); err != nil {
			t.Fatal(err)
		}
	}
	if err := hub.Register(ctx, Device{UserID: 1, Kind: KindWeb, Token: "x"}); err != ErrUnsupported {
		t.Fatalf("web push without VAPID keys: %v", err)
	}

	hub.Notify(ctx, 1, Notification{Title: "Bulle", Body: "Morning brief is ready", Path: "sessions/s1"})

	if len(got) != 2 || !strings.Contains(got[0], `"path":"sessions/s1"`) ||
		!strings.Contains(got[0], `"channel_id":"bulle"`) || !strings.Contains(got[0], `"tag":"sessions/s1"`) {
		t.Fatalf("fcm requests: %q", got)
	}
	left, _ := devices.Devices(ctx, 1)
	if len(left) != 1 || left[0].Token != "live" {
		t.Fatalf("devices after delivery: %+v", left)
	}
}

func TestLiveDeliversToSubscribersOnly(t *testing.T) {
	hub := NewHub(HubConfig{})
	mine, stop := hub.Subscribe(7)
	other, stopOther := hub.Subscribe(8)
	defer stopOther()
	hub.Notify(context.Background(), 7, Notification{Title: "Bulle", Body: "Approval needed", Path: "sessions/s1"})
	select {
	case e := <-mine:
		if e.Type != EventNotification || !strings.Contains(string(e.Data), `"path":"sessions/s1"`) {
			t.Fatalf("got %s %s", e.Type, e.Data)
		}
	default:
		t.Fatal("the subscriber got nothing")
	}
	select {
	case e := <-other:
		t.Fatalf("another user got %s", e.Data)
	default:
	}
	stop()
	for range 20 {
		hub.Publish(7, "session.updated", map[string]string{"session_id": "s1"})
	}
	if len(mine) != 0 {
		t.Fatal("a stopped subscriber still receives")
	}
}

func TestLiveCutsOffASlowListener(t *testing.T) {
	hub := NewHub(HubConfig{})
	slow, stop := hub.Subscribe(8)
	defer stop()
	for range liveBuffer + 1 {
		hub.Publish(8, "session.updated", map[string]string{"session_id": "s1"})
	}
	received := 0
	for range slow {
		received++
	}
	if received != liveBuffer {
		t.Fatalf("a full listener should be closed after its buffer, got %d events", received)
	}
	var nilHub *Hub
	nilHub.Publish(8, "session.updated", nil)
}
