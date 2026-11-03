package push

import (
	"context"
	"errors"
	"log/slog"
)

// ErrUnsupported means the server cannot deliver to that kind of device.
var ErrUnsupported = errors.New("this server cannot send to that kind of device")

// HubConfig lists the delivery channels. Any of them may be nil.
type HubConfig struct {
	Ntfy    *Sender
	FCM     *FCM
	Web     *WebPush
	Devices DeviceStore
	Live    *Live
	Logger  *slog.Logger
}

// Hub fans one notification out to ntfy and every device of the user.
// A nil Hub is a safe no-op.
type Hub struct{ cfg HubConfig }

// NewHub returns a Hub over the configured channels.
func NewHub(cfg HubConfig) *Hub {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.Live == nil {
		cfg.Live = NewLive()
	}
	return &Hub{cfg: cfg}
}

// Ntfy returns the ntfy sender, nil when ntfy is off.
func (h *Hub) Ntfy() *Sender {
	if h == nil {
		return nil
	}
	return h.cfg.Ntfy
}

// Native reports whether native Android notifications are on.
func (h *Hub) Native() bool { return h != nil && h.cfg.FCM != nil && h.cfg.Devices != nil }

// WebPushKey returns the VAPID public key browsers subscribe with, or "".
func (h *Hub) WebPushKey() string {
	if h == nil || h.cfg.Web == nil || h.cfg.Devices == nil {
		return ""
	}
	return h.cfg.Web.publicKey
}

// Enabled reports whether any channel delivers notifications. The live
// stream to desktop apps is always on.
func (h *Hub) Enabled() bool { return h != nil }

// Subscribe streams the user's notifications to an open desktop app.
func (h *Hub) Subscribe(userID int64) (<-chan Notification, func()) {
	return h.cfg.Live.Subscribe(userID)
}

// Register records a device when the server can send to its kind.
func (h *Hub) Register(ctx context.Context, d Device) error {
	if (d.Kind == KindFCM && !h.Native()) || (d.Kind == KindWeb && h.WebPushKey() == "") {
		return ErrUnsupported
	}
	if d.Kind != KindFCM && d.Kind != KindWeb {
		return ErrUnsupported
	}
	return h.cfg.Devices.SaveDevice(ctx, d)
}

// Unregister forgets one of the user's devices.
func (h *Hub) Unregister(ctx context.Context, userID int64, token string) error {
	if h == nil || h.cfg.Devices == nil {
		return nil
	}
	return h.cfg.Devices.DeleteDevice(ctx, userID, token)
}

// Notify delivers to every channel. Failures are logged and never returned;
// a device the push service reports as gone is forgotten.
func (h *Hub) Notify(ctx context.Context, userID int64, n Notification) {
	if h == nil {
		return
	}
	h.cfg.Live.publish(userID, n)
	h.cfg.Ntfy.Notify(ctx, userID, n)
	if h.cfg.Devices == nil || (h.cfg.FCM == nil && h.cfg.Web == nil) {
		return
	}
	devices, err := h.cfg.Devices.Devices(ctx, userID)
	if err != nil {
		h.cfg.Logger.Warn("listing push devices failed", slog.Any("error", err))
		return
	}
	for _, d := range devices {
		h.deliver(ctx, d, n)
	}
}

func (h *Hub) deliver(ctx context.Context, d Device, n Notification) {
	var err error
	switch {
	case d.Kind == KindFCM && h.cfg.FCM != nil:
		err = h.cfg.FCM.send(ctx, d.Token, n)
	case d.Kind == KindWeb && h.cfg.Web != nil:
		err = h.cfg.Web.send(ctx, d, n)
	default:
		return
	}
	if errors.Is(err, ErrGone) {
		_ = h.cfg.Devices.DeleteDevice(ctx, d.UserID, d.Token)
		return
	}
	if err != nil {
		h.cfg.Logger.Warn("push delivery failed", slog.String("kind", d.Kind), slog.Any("error", err))
	}
}
