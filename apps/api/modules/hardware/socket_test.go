package hardware_test

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	internal "github.com/FacileStudio/Bulle/apps/api/internal/hardware"
	"github.com/FacileStudio/Bulle/apps/api/modules/hardware"
	"github.com/coder/websocket"
)

func TestSocketRefusesAMissingSubprotocol(t *testing.T) {
	srv := server(t, newFakeStore())
	claimed := claim(t, srv)
	_, res, err := websocket.Dial(context.Background(), wsURL(srv), &websocket.DialOptions{
		HTTPHeader: http.Header{"Authorization": {"Bearer " + claimed.DeviceToken}},
	})
	if err == nil || res == nil || res.StatusCode != http.StatusBadRequest {
		t.Fatalf("socket without bohs.v1: %v", err)
	}
}

func TestSocketRefusesAnUnknownToken(t *testing.T) {
	srv := server(t, newFakeStore())
	unknown, _ := internal.MintToken()
	_, res, err := websocket.Dial(context.Background(), wsURL(srv), &websocket.DialOptions{
		Subprotocols: []string{"bohs.v1"},
		HTTPHeader:   http.Header{"Authorization": {"Bearer " + unknown}},
	})
	if err == nil || res == nil || res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("socket with an unknown token: %v", err)
	}
}

func TestRelayRefusesAnotherOrigin(t *testing.T) {
	srv := server(t, newFakeStore())
	claimed := claim(t, srv)
	_, res, err := websocket.Dial(context.Background(), wsURL(srv)+"?device_id="+claimed.Device.ID, &websocket.DialOptions{
		Subprotocols: []string{"bohs.v1"},
		HTTPHeader:   http.Header{"X-User": {"1"}, "Origin": {"https://evil.example"}},
	})
	if err == nil || res == nil || res.StatusCode != http.StatusForbidden {
		t.Fatalf("cross-origin relay: %v", err)
	}
}

func TestRelayRecordsItsTransport(t *testing.T) {
	st := newFakeStore()
	srv := server(t, st)
	claimed := claim(t, srv)
	url := wsURL(srv) + "?device_id=" + claimed.Device.ID
	_, res, err := websocket.Dial(context.Background(), url+"&transport=wifi", &websocket.DialOptions{
		Subprotocols: []string{"bohs.v1"},
		HTTPHeader:   http.Header{"X-User": {"1"}},
	})
	if err == nil || res == nil || res.StatusCode != http.StatusBadRequest {
		t.Fatalf("relay with transport=wifi: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, url+"&transport=usb", &websocket.DialOptions{
		Subprotocols: []string{"bohs.v1"},
		HTTPHeader:   http.Header{"X-User": {"1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()
	if _, _, err := conn.Read(ctx); err != nil {
		t.Fatal(err)
	}
	if got := st.lastTransport(); got != internal.TransportUSB {
		t.Fatalf("recorded transport = %q, want usb", got)
	}
}

func TestSocketClosesOnAnOversizeFrame(t *testing.T) {
	conn, ctx := openDirect(t, newFakeStore(), nil)
	if err := conn.Write(ctx, websocket.MessageBinary, control(1024)); err != nil {
		t.Fatal(err)
	}
	if err := conn.Write(ctx, websocket.MessageBinary, control(1025)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := conn.Read(ctx); websocket.CloseStatus(err) != websocket.StatusMessageTooBig {
		t.Fatalf("read after a 1025-byte frame = %v, want close 1009", err)
	}
}

func TestSocketClosesOnAProvisioningFrame(t *testing.T) {
	conn, ctx := openDirect(t, newFakeStore(), nil)
	if err := conn.Write(ctx, websocket.MessageBinary, []byte{0x04, 0x00}); err != nil {
		t.Fatal(err)
	}
	if err := conn.Write(ctx, websocket.MessageBinary, []byte{0x03, '{', '}'}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := conn.Read(ctx); websocket.CloseStatus(err) != websocket.StatusPolicyViolation {
		t.Fatalf("read after a 0x03 frame = %v, want close 1008", err)
	}
}

func TestRevokeDuringTheHandshakeStillClosesTheSocket(t *testing.T) {
	st := newFakeStore()
	conn, ctx := openDirect(t, st, func(id string) { st.forget(id) })
	if _, _, err := conn.Read(ctx); websocket.CloseStatus(err) != internal.StatusRevoked {
		t.Fatalf("first read on a socket revoked mid-handshake = %v, want close %d", err, internal.StatusRevoked)
	}
}

func TestRelayClosesWhenTheSessionEnds(t *testing.T) {
	st := newFakeStore()
	signedOut := &atomic.Bool{}
	srv := serverWith(t, hardware.Config{Devices: st, Sockets: st, Keepalive: 20 * time.Millisecond}, signedOut)
	claimed := claim(t, srv)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, wsURL(srv)+"?device_id="+claimed.Device.ID, &websocket.DialOptions{
		Subprotocols: []string{"bohs.v1"},
		HTTPHeader:   http.Header{"X-User": {"1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()
	if _, _, err := conn.Read(ctx); err != nil {
		t.Fatal(err)
	}
	signedOut.Store(true)
	if _, _, err := conn.Read(ctx); websocket.CloseStatus(err) != internal.StatusRevoked {
		t.Fatalf("read after sign-out = %v, want close %d", err, internal.StatusRevoked)
	}
}

func openDirect(t *testing.T, st *fakeStore, afterLookup func(id string)) (*websocket.Conn, context.Context) {
	t.Helper()
	srv := server(t, st)
	claimed := claim(t, srv)
	if afterLookup != nil {
		st.afterLookup = func() { afterLookup(claimed.Device.ID) }
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	conn, _, err := websocket.Dial(ctx, wsURL(srv), &websocket.DialOptions{
		Subprotocols: []string{"bohs.v1"},
		HTTPHeader:   http.Header{"Authorization": {"Bearer " + claimed.DeviceToken}},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.CloseNow() })
	if afterLookup == nil {
		if _, _, err := conn.Read(ctx); err != nil {
			t.Fatal(err)
		}
	}
	return conn, ctx
}

func control(size int) []byte {
	frame := make([]byte, size)
	frame[0] = 0x02
	return frame
}
