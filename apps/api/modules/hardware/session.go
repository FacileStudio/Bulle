package hardware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	internal "github.com/FacileStudio/Bulle/apps/api/internal/hardware"
	"github.com/coder/websocket"
)

const (
	opcodeControl   byte = 0x02
	opcodeProvision byte = 0x03
)

const (
	maxFrameBytes = 1024
	pingTimeout   = 10 * time.Second
)

var toolsList = append([]byte{opcodeControl}, `{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`...)

type session struct {
	device    internal.Device
	transport string
	check     credentialCheck
}

func (h *Handler) serve(w http.ResponseWriter, r *http.Request, opts *websocket.AcceptOptions, s session) {
	device, transport, check := s.device, s.transport, s.check
	conn, err := websocket.Accept(w, r, opts)
	if err != nil {
		h.cfg.Logger.Debug("hardware socket refused", slog.String("device_id", device.ID), slog.Any("error", err))
		return
	}
	defer conn.CloseNow()
	conn.SetReadLimit(maxFrameBytes)
	untrack := h.cfg.Registry.Add(device.ID, conn)
	defer untrack()
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	if err := check(ctx); err != nil {
		if !h.closeFor(conn, device.ID, err) {
			_ = conn.Close(websocket.StatusInternalError, "could not check the device credential")
		}
		return
	}
	h.touch(ctx, device.ID, transport)
	defer h.touch(context.WithoutCancel(ctx), device.ID, transport)
	if err := conn.Write(ctx, websocket.MessageBinary, toolsList); err != nil {
		return
	}
	go h.keepalive(ctx, conn, device.ID, check)
	h.read(ctx, conn, device.ID)
}

func (h *Handler) read(ctx context.Context, conn *websocket.Conn, deviceID string) {
	for {
		kind, frame, err := conn.Read(ctx)
		if err != nil {
			return
		}
		if kind != websocket.MessageBinary || len(frame) == 0 {
			_ = conn.Close(websocket.StatusUnsupportedData, "bohs.v1 carries binary frames only")
			return
		}
		if frame[0] == opcodeProvision {
			_ = conn.Close(websocket.StatusPolicyViolation, "provisioning frames are BLE only")
			return
		}
		if frame[0] > opcodeProvision {
			continue
		}
		h.cfg.Logger.Debug("hardware frame",
			slog.String("device_id", deviceID), slog.Int("opcode", int(frame[0])), slog.Int("bytes", len(frame)))
	}
}

func (h *Handler) keepalive(ctx context.Context, conn *websocket.Conn, deviceID string, check credentialCheck) {
	ticker := time.NewTicker(h.cfg.Keepalive)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !h.stillOpen(ctx, conn, deviceID, check) {
				return
			}
		}
	}
}

func (h *Handler) stillOpen(ctx context.Context, conn *websocket.Conn, deviceID string, check credentialCheck) bool {
	if !ping(ctx, conn) {
		return false
	}
	err := check(ctx)
	switch {
	case ctx.Err() != nil:
		return false
	case err == nil:
		return true
	default:
		return !h.closeFor(conn, deviceID, err)
	}
}

func ping(ctx context.Context, conn *websocket.Conn) bool {
	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()
	if err := conn.Ping(pingCtx); err != nil {
		conn.CloseNow()
		return false
	}
	return true
}

func (h *Handler) closeFor(conn *websocket.Conn, deviceID string, err error) bool {
	if errors.Is(err, errRevoked) {
		_ = conn.Close(internal.StatusRevoked, "device revoked or session ended")
		return true
	}
	h.cfg.Logger.Warn("failed to check the device credential", slog.String("device_id", deviceID), slog.Any("error", err))
	return false
}

func (h *Handler) touch(ctx context.Context, deviceID, transport string) {
	if err := h.cfg.Sockets.Touch(ctx, deviceID, transport, time.Now().UTC()); err != nil {
		h.cfg.Logger.Warn("failed to record the device connection",
			slog.String("device_id", deviceID), slog.Any("error", err))
	}
}
