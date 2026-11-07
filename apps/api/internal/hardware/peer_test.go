package hardware_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/hardware"
	"github.com/coder/websocket"
)

type echoDevice struct {
	peer  *hardware.Peer
	sent  chan map[string]any
	reply func(req map[string]any) string
}

func (d *echoDevice) Write(_ context.Context, _ websocket.MessageType, p []byte) error {
	var req map[string]any
	if p[0] != hardware.OpcodeControl || json.Unmarshal(p[1:], &req) != nil {
		return fmt.Errorf("bad frame %q", p)
	}
	d.sent <- req
	if d.reply != nil {
		go d.peer.Deliver([]byte(d.reply(req)))
	}
	return nil
}

func (d *echoDevice) Close(websocket.StatusCode, string) error { return nil }

func switchDevice(t *testing.T, registry *hardware.Registry, name string, userID int64) *echoDevice {
	t.Helper()
	dev := &echoDevice{sent: make(chan map[string]any, 8)}
	dev.reply = func(req map[string]any) string {
		id, _ := json.Marshal(req["id"])
		switch req["method"] {
		case "tools/list":
			return `{"jsonrpc":"2.0","id":` + string(id) + `,"result":{"tools":[{"name":"switch.set","description":"Turns the light on or off","inputSchema":{"type":"object","properties":{"state":{"type":"string","enum":["on","off"]}},"required":["state"]}}]}}`
		case "tools/call":
			args := req["params"].(map[string]any)["arguments"].(map[string]any)
			if args["state"] == "broken" {
				return `{"jsonrpc":"2.0","id":` + string(id) + `,"error":{"code":-32602,"message":"invalid params"}}`
			}
			return `{"jsonrpc":"2.0","id":` + string(id) + `,"result":{"content":[{"type":"text","text":"light is ` + args["state"].(string) + `"}],"isError":false}}`
		}
		return `{"jsonrpc":"2.0","id":` + string(id) + `,"error":{"code":-32601,"message":"method not found"}}`
	}
	dev.peer = hardware.NewPeer(hardware.Device{ID: "hw_0123456789abcdef0123456789abcdef", UserID: userID, Name: name}, hardware.TransportWiFi, dev)
	registry.Add(dev.peer.Device.ID, dev.peer)
	if err := dev.peer.Discover(context.Background()); err != nil {
		t.Fatalf("discover: %v", err)
	}
	return dev
}

func TestAgentToolsDriveTheDevice(t *testing.T) {
	registry := hardware.NewRegistry()
	switchDevice(t, registry, "Living Room Light", 7)
	if tools, _ := registry.AgentTools(8); len(tools) != 0 {
		t.Fatalf("another user sees %d device tools", len(tools))
	}
	tools, note := registry.AgentTools(7)
	if len(tools) != 1 || tools[0].Name() != "living_room_light__switch_set" {
		t.Fatalf("tools = %v", tools)
	}
	if !strings.Contains(note, "Living Room Light") || !strings.Contains(tools[0].Description(), "Living Room Light") {
		t.Fatalf("the device name is missing from the note %q or the description %q", note, tools[0].Description())
	}
	out, err := tools[0].Run(context.Background(), json.RawMessage(`{"state":"on"}`))
	if err != nil || out != "light is on" {
		t.Fatalf("run = %q, %v", out, err)
	}
	if _, err := tools[0].Run(context.Background(), json.RawMessage(`{"state":"broken"}`)); err == nil || !strings.Contains(err.Error(), "invalid params") {
		t.Fatalf("device error = %v", err)
	}
}

func TestPeerCallFailsWhenTheSocketCloses(t *testing.T) {
	dev := &echoDevice{sent: make(chan map[string]any, 1)}
	dev.peer = hardware.NewPeer(hardware.Device{ID: "hw_x"}, hardware.TransportWiFi, dev)
	done := make(chan error, 1)
	go func() {
		_, err := dev.peer.Call(context.Background(), "tools/call", nil)
		done <- err
	}()
	<-dev.sent
	dev.peer.Shutdown()
	select {
	case err := <-done:
		if err != hardware.ErrPeerClosed {
			t.Fatalf("err = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("the call kept waiting on a closed socket")
	}
	if dev.peer.Deliver([]byte(`{"jsonrpc":"2.0","id":1,"result":{}}`)) {
		t.Fatal("a late answer matched a forgotten call")
	}
}

func TestAgentToolsDropAToolWhoseSchemaAModelWouldRefuse(t *testing.T) {
	registry := hardware.NewRegistry()
	dev := &echoDevice{sent: make(chan map[string]any, 8)}
	dev.reply = func(req map[string]any) string {
		id, _ := json.Marshal(req["id"])
		long := strings.Repeat("x ", 400)
		return `{"jsonrpc":"2.0","id":` + string(id) + `,"result":{"tools":[` +
			`{"name":"bad","description":"d","inputSchema":{"type":"object","properties":{"x":{"type":"banana"}}}},` +
			`{"name":"good","description":"line one\nline two ` + long + `","inputSchema":{"type":"object","properties":{}}}]}}`
	}
	dev.peer = hardware.NewPeer(hardware.Device{ID: "hw_0123456789abcdef0123456789abcdef", UserID: 1, Name: "Lamp"}, hardware.TransportWiFi, dev)
	registry.Add(dev.peer.Device.ID, dev.peer)
	if err := dev.peer.Discover(context.Background()); err != nil {
		t.Fatal(err)
	}
	tools, _ := registry.AgentTools(1)
	if len(tools) != 1 || tools[0].Name() != "lamp__good" {
		t.Fatalf("tools = %v", tools)
	}
	if desc := tools[0].Description(); strings.Contains(desc, "\n") || len(desc) > 260 {
		t.Fatalf("description not capped to one short line: %q", desc)
	}
}
