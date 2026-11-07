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

const opcodeProvision byte = 0x03

const (
	pingTimeout     = 10 * time.Second
	discoverTimeout = 10 * time.Second
)

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
	conn.SetReadLimit(internal.MaxFrameBytes)
	peer := internal.NewPeer(device, transport, conn)
	defer peer.Shutdown()
	untrack := h.cfg.Registry.Add(device.ID, peer)
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
	go h.keepalive(ctx, conn, device.ID, check)
	go h.discover(ctx, peer)
	h.read(ctx, conn, peer)
}

func (h *Handler) discover(ctx context.Context, peer *internal.Peer) {
	ctx, cancel := context.WithTimeout(ctx, discoverTimeout)
	defer cancel()
	if err := peer.Discover(ctx); err != nil {
		h.cfg.Logger.Debug("hardware tools/list failed", slog.String("device_id", peer.Device.ID), slog.Any("error", err))
	}
}

func (h *Handler) read(ctx context.Context, conn *websocket.Conn, peer *internal.Peer) {
	deviceID := peer.Device.ID
	for {
		kind, frame, err := conn.Read(ctx)
		if err != nil {
			return
		}
		if !accept(conn, kind, frame) {
			return
		}
		if frame[0] > opcodeProvision || frame[0] == internal.OpcodeControl && peer.Deliver(frame[1:]) {
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
