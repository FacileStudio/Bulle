package hardware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	internal "github.com/FacileStudio/Bulle/apps/api/internal/hardware"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/coder/websocket"
)

const subprotocol = "bohs.v1"

var relayTransports = map[string]string{
	"":                    internal.TransportBLE,
	internal.TransportBLE: internal.TransportBLE,
	internal.TransportUSB: internal.TransportUSB,
}

// socket opens a bohs.v1 gateway socket. A device on Wi-Fi presents its own
// token as a bearer; anything else is the signed-in app relaying a device it
// owns over BLE or USB, authenticated like every other route.
func (h *Handler) socket(w http.ResponseWriter, r *http.Request) {
	if !offersSubprotocol(r) {
		httpjson.WriteError(w, troncerrors.Invalid("the "+subprotocol+" subprotocol is required"))
		return
	}
	if token, ok := bearerToken(r); ok {
		h.direct(w, r, token)
		return
	}
	h.requireAuth(http.HandlerFunc(h.relay)).ServeHTTP(w, r)
}

func (h *Handler) direct(w http.ResponseWriter, r *http.Request, token string) {
	hash := internal.HashToken(token)
	device, err := h.cfg.Sockets.ByTokenHash(r.Context(), hash)
	if errors.Is(err, internal.ErrNotFound) {
		httpjson.WriteError(w, troncerrors.Unauthorized("invalid device token"))
		return
	}
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to read the device", err))
		return
	}
	h.serve(w, r, &websocket.AcceptOptions{
		Subprotocols:       []string{subprotocol},
		InsecureSkipVerify: true,
	}, session{device: device, transport: internal.TransportWiFi, check: func(ctx context.Context) error {
		return h.tokenValid(ctx, hash, device.ID)
	}})
}

func (h *Handler) relay(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	transport, ok := relayTransports[r.URL.Query().Get("transport")]
	if !ok {
		httpjson.WriteError(w, troncerrors.Invalid("transport must be ble or usb"))
		return
	}
	device, err := h.cfg.Sockets.Owned(r.Context(), identity.UserID, r.URL.Query().Get("device_id"))
	if errors.Is(err, internal.ErrNotFound) {
		httpjson.WriteError(w, troncerrors.NotFound("device not found"))
		return
	}
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to read the device", err))
		return
	}
	h.serve(w, r, &websocket.AcceptOptions{
		Subprotocols:   []string{subprotocol},
		OriginPatterns: h.cfg.Origins,
	}, session{device: device, transport: transport, check: func(ctx context.Context) error {
		return h.relayValid(ctx, r, identity.UserID, device.ID)
	}})
}

func bearerToken(r *http.Request) (string, bool) {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || !internal.IsToken(token) {
		return "", false
	}
	return token, true
}

func offersSubprotocol(r *http.Request) bool {
	for _, header := range r.Header.Values("Sec-WebSocket-Protocol") {
		for offered := range strings.SplitSeq(header, ",") {
			if strings.TrimSpace(offered) == subprotocol {
				return true
			}
		}
	}
	return false
}
