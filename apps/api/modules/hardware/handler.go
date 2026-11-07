package hardware

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	internal "github.com/FacileStudio/Bulle/apps/api/internal/hardware"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

// Devices is the part of the device store the pairing routes use.
type Devices interface {
	Claim(ctx context.Context, d internal.Device, tokenHash string) (internal.Device, error)
	Devices(ctx context.Context, userID int64) ([]internal.Device, error)
	Revoke(ctx context.Context, userID int64, id string) error
}

// Sockets is the part of the device store the gateway socket uses.
type Sockets interface {
	Owned(ctx context.Context, userID int64, id string) (internal.Device, error)
	ByTokenHash(ctx context.Context, hash string) (internal.Device, error)
	Touch(ctx context.Context, id, transport string, at time.Time) error
}

// Config wires the hardware routes. Origins lists the extra browser origins
// allowed to open the relay socket, on top of the API's own host. GatewayURL
// is the socket URL handed to devices; empty means wss:// on the request host.
// Keepalive is how often a socket is pinged and its credential checked again;
// zero means 30 seconds.
type Config struct {
	Devices    Devices
	Sockets    Sockets
	Registry   *internal.Registry
	Origins    []string
	GatewayURL string
	Pairings   *internal.Pairings
	Keepalive  time.Duration
	Logger     *slog.Logger
}

// Handler serves the hardware routes.
type Handler struct {
	cfg         Config
	requireAuth func(http.Handler) http.Handler
}

func (h *Handler) claim(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	var body ClaimRequest
	if err := httpjson.DecodeJSON(w, r, &body); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	device, err := newDevice(identity.UserID, body)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	token, hash := internal.MintToken()
	claimed, err := h.cfg.Devices.Claim(r.Context(), device, hash)
	if err != nil {
		h.claimFailed(w, err)
		return
	}
	h.cfg.Registry.Close(claimed.ID)
	httpjson.WriteJSON(w, http.StatusCreated, ClaimResponse{
		Device:       toResponse(claimed),
		DeviceToken:  token,
		GatewayWSURL: h.gatewayURL(r),
	})
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	devices, err := h.cfg.Devices.Devices(r.Context(), identity.UserID)
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to list devices", err))
		return
	}
	out := make([]DeviceResponse, 0, len(devices))
	for _, device := range devices {
		out = append(out, toResponse(device))
	}
	httpjson.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) revoke(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	id := chi.URLParam(r, "id")
	err := h.cfg.Devices.Revoke(r.Context(), identity.UserID, id)
	if errors.Is(err, internal.ErrNotFound) {
		httpjson.WriteError(w, troncerrors.NotFound("device not found"))
		return
	}
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to revoke the device", err))
		return
	}
	h.cfg.Registry.Close(id)
	w.WriteHeader(http.StatusNoContent)
}

func toResponse(d internal.Device) DeviceResponse {
	capabilities := json.RawMessage(d.Capabilities)
	if !json.Valid(capabilities) {
		capabilities = json.RawMessage("[]")
	}
	return DeviceResponse{
		ID:            d.ID,
		Name:          d.Name,
		DeviceClass:   d.DeviceClass,
		HardwareID:    d.HardwareID,
		Capabilities:  capabilities,
		LastSeenAt:    d.LastSeenAt,
		LastTransport: d.LastTransport,
		CreatedAt:     d.CreatedAt,
	}
}

func (h *Handler) gatewayURL(r *http.Request) string {
	if h.cfg.GatewayURL != "" {
		return h.cfg.GatewayURL
	}
	return "wss://" + r.Host + "/api/hardware/ws"
}
