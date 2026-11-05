package hardware

import (
	"slices"
	"sync"

	"github.com/coder/websocket"
)

// StatusRevoked is the WebSocket close code a socket receives when its device
// is revoked, its token replaced, or the session that relayed it ended.
const StatusRevoked websocket.StatusCode = 4401

// MaxSocketsPerDevice caps the live sockets of one device: one direct and one
// relayed. A newer socket past the cap replaces the oldest.
const MaxSocketsPerDevice = 2

// Closer is the part of a live socket the registry needs.
type Closer interface {
	Close(code websocket.StatusCode, reason string) error
}

// Registry tracks the live /api/hardware/ws sockets of each device, so a
// revoke can cut them.
type Registry struct {
	mu    sync.Mutex
	conns map[string][]Closer
}

// NewRegistry answers an empty registry.
func NewRegistry() *Registry {
	return &Registry{conns: map[string][]Closer{}}
}

// Add tracks conn under the device id and answers the function that stops
// tracking it. Past MaxSocketsPerDevice it closes the oldest socket of the
// device with StatusPolicyViolation.
func (r *Registry) Add(deviceID string, conn Closer) func() {
	r.mu.Lock()
	conns := append(r.conns[deviceID], conn)
	var evicted []Closer
	if over := len(conns) - MaxSocketsPerDevice; over > 0 {
		evicted, conns = slices.Clone(conns[:over]), slices.Clone(conns[over:])
	}
	r.conns[deviceID] = conns
	r.mu.Unlock()
	for _, old := range evicted {
		go func() { _ = old.Close(websocket.StatusPolicyViolation, "replaced by a newer session") }()
	}
	return func() { r.remove(deviceID, conn) }
}

// Close shuts every live socket of the device with StatusRevoked and answers
// how many it closed. The close handshakes finish in the background.
func (r *Registry) Close(deviceID string) int {
	r.mu.Lock()
	conns := r.conns[deviceID]
	delete(r.conns, deviceID)
	r.mu.Unlock()
	for _, conn := range conns {
		go func() { _ = conn.Close(StatusRevoked, "device revoked") }()
	}
	return len(conns)
}

func (r *Registry) remove(deviceID string, conn Closer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	conns := slices.DeleteFunc(slices.Clone(r.conns[deviceID]), func(c Closer) bool { return c == conn })
	if len(conns) == 0 {
		delete(r.conns, deviceID)
		return
	}
	r.conns[deviceID] = conns
}
