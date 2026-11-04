package hardware

import (
	"sync"

	"github.com/coder/websocket"
)

// StatusRevoked is the WebSocket close code a socket receives when its device
// is revoked or its token replaced.
const StatusRevoked websocket.StatusCode = 4401

// Closer is the part of a live socket the registry needs.
type Closer interface {
	Close(code websocket.StatusCode, reason string) error
}

// Registry tracks the live /api/hardware/ws sockets of each device, so a
// revoke can cut them.
type Registry struct {
	mu    sync.Mutex
	conns map[string]map[Closer]struct{}
}

// NewRegistry answers an empty registry.
func NewRegistry() *Registry {
	return &Registry{conns: map[string]map[Closer]struct{}{}}
}

// Add tracks conn under the device id and answers the function that stops
// tracking it.
func (r *Registry) Add(deviceID string, conn Closer) func() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.conns[deviceID] == nil {
		r.conns[deviceID] = map[Closer]struct{}{}
	}
	r.conns[deviceID][conn] = struct{}{}
	return func() { r.remove(deviceID, conn) }
}

// Close shuts every live socket of the device with StatusRevoked and answers
// how many it closed. The close handshakes finish in the background.
func (r *Registry) Close(deviceID string) int {
	r.mu.Lock()
	conns := r.conns[deviceID]
	delete(r.conns, deviceID)
	r.mu.Unlock()
	for conn := range conns {
		go func() { _ = conn.Close(StatusRevoked, "device revoked") }()
	}
	return len(conns)
}

func (r *Registry) remove(deviceID string, conn Closer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.conns[deviceID], conn)
	if len(r.conns[deviceID]) == 0 {
		delete(r.conns, deviceID)
	}
}
