package hardware

import (
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

// socket opens a bohs.v1 gateway socket. A device on Wi-Fi presents its own
// token as a bearer; anything else is the signed-in app relaying a device it
// owns over BLE, authenticated like every other route.
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
	device, err := h.cfg.Sockets.ByTokenHash(r.Context(), internal.HashToken(token))
	if errors.Is(err, internal.ErrNotFound) {
		httpjson.WriteError(w, troncerrors.Unauthorized("invalid device token"))
		return
	}
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to read the device", err))
		return
	}
	h.serve(w, r, device, internal.TransportWiFi, &websocket.AcceptOptions{
		Subprotocols:       []string{subprotocol},
		InsecureSkipVerify: true,
	})
}

func (h *Handler) relay(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
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
	h.serve(w, r, device, internal.TransportBLE, &websocket.AcceptOptions{
		Subprotocols:   []string{subprotocol},
		OriginPatterns: h.cfg.Origins,
	})
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
