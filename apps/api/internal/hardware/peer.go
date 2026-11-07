package hardware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/coder/websocket"
)

// OpcodeControl starts every bohs.v1 JSON-RPC frame.
const OpcodeControl byte = 0x02

// MaxFrameBytes is the largest bohs.v1 frame, opcode included.
const MaxFrameBytes = 1024

// ErrPeerClosed reports a call on a socket that has gone away.
var ErrPeerClosed = errors.New("hardware: the device disconnected")

// ErrFrameTooLarge reports a request that does not fit in one bohs.v1 frame.
var ErrFrameTooLarge = errors.New("hardware: the request does not fit in one frame")

// RPCError is a JSON-RPC error object a device answered.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("device error %d: %s", e.Code, e.Message)
}

// Sender writes one binary frame to a live socket.
type Sender interface {
	Write(ctx context.Context, typ websocket.MessageType, p []byte) error
	Close(code websocket.StatusCode, reason string) error
}

type reply struct {
	result json.RawMessage
	err    error
}

// Peer is one live gateway socket seen as a Reverse MCP server: the gateway
// sends it JSON-RPC requests and matches the answers by id.
type Peer struct {
	Device    Device
	Transport string
	conn      Sender

	mu      sync.Mutex
	next    int64
	pending map[string]chan reply
	tools   []EdgeTool
	ready   bool
	closed  bool
}

// NewPeer wraps a socket of device.
func NewPeer(device Device, transport string, conn Sender) *Peer {
	return &Peer{Device: device, Transport: transport, conn: conn, pending: map[string]chan reply{}}
}

// Close closes the socket and fails every call still waiting.
func (p *Peer) Close(code websocket.StatusCode, reason string) error {
	p.Shutdown()
	return p.conn.Close(code, reason)
}

// Shutdown fails every waiting call and refuses new ones, without touching
// the socket.
func (p *Peer) Shutdown() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	for id, ch := range p.pending {
		ch <- reply{err: ErrPeerClosed}
		delete(p.pending, id)
	}
}

// Call sends one request and waits for its answer, the context or the socket.
func (p *Peer) Call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	ch := make(chan reply, 1)
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, ErrPeerClosed
	}
	p.next++
	id := strconv.FormatInt(p.next, 10)
	p.pending[id] = ch
	p.mu.Unlock()
	defer p.forget(id)
	frame, err := requestFrame(p.next, method, params)
	if err != nil {
		return nil, err
	}
	if err := p.conn.Write(ctx, websocket.MessageBinary, frame); err != nil {
		return nil, ErrPeerClosed
	}
	select {
	case r := <-ch:
		return r.result, r.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *Peer) forget(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.pending, id)
}

func requestFrame(id int64, method string, params any) ([]byte, error) {
	if params == nil {
		params = map[string]any{}
	}
	body, err := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params})
	if err != nil {
		return nil, err
	}
	if len(body)+1 > MaxFrameBytes {
		return nil, ErrFrameTooLarge
	}
	return append([]byte{OpcodeControl}, body...), nil
}

// Deliver hands a 0x02 payload from the device to the call waiting on its id.
// It answers false for anything that is not an answer to a pending call, such
// as a device notification.
func (p *Peer) Deliver(payload []byte) bool {
	var msg struct {
		ID     json.RawMessage `json:"id"`
		Method string          `json:"method"`
		Result json.RawMessage `json:"result"`
		Error  *RPCError       `json:"error"`
	}
	if err := json.Unmarshal(payload, &msg); err != nil || msg.Method != "" || len(msg.ID) == 0 {
		return false
	}
	id := string(msg.ID)
	if unquoted, err := strconv.Unquote(id); err == nil {
		id = unquoted
	}
	p.mu.Lock()
	ch, ok := p.pending[id]
	delete(p.pending, id)
	p.mu.Unlock()
	if !ok {
		return false
	}
	if msg.Error != nil {
		ch <- reply{err: msg.Error}
		return true
	}
	ch <- reply{result: msg.Result}
	return true
}
