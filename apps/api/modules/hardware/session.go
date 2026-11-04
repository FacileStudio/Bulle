package hardware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	internal "github.com/FacileStudio/Bulle/apps/api/internal/hardware"
	"github.com/coder/websocket"
)

const opcodeControl byte = 0x02

const (
	maxFrameBytes = 64 << 10
	pingInterval  = 30 * time.Second
	pingTimeout   = 10 * time.Second
)

var toolsList = append([]byte{opcodeControl}, `{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`...)

func (h *Handler) serve(w http.ResponseWriter, r *http.Request, device internal.Device, transport string, opts *websocket.AcceptOptions) {
	conn, err := websocket.Accept(w, r, opts)
	if err != nil {
		return
	}
	defer conn.CloseNow()
	conn.SetReadLimit(maxFrameBytes)
	untrack := h.cfg.Registry.Add(device.ID, conn)
	defer untrack()
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	h.touch(ctx, device.ID, transport)
	defer h.touch(context.WithoutCancel(ctx), device.ID, transport)
	if err := conn.Write(ctx, websocket.MessageBinary, toolsList); err != nil {
		return
	}
	go keepalive(ctx, conn)
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
		h.cfg.Logger.Debug("hardware frame",
			slog.String("device_id", deviceID), slog.Int("opcode", int(frame[0])), slog.Int("bytes", len(frame)))
	}
}

func keepalive(ctx context.Context, conn *websocket.Conn) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
			err := conn.Ping(pingCtx)
			cancel()
			if err != nil {
				conn.CloseNow()
				return
			}
		}
	}
}

func (h *Handler) touch(ctx context.Context, deviceID, transport string) {
	if err := h.cfg.Sockets.Touch(ctx, deviceID, transport, time.Now().UTC()); err != nil {
		h.cfg.Logger.Warn("failed to record the device connection",
			slog.String("device_id", deviceID), slog.Any("error", err))
	}
}
