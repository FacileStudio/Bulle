package hardware

import (
	"cmp"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	internal "github.com/FacileStudio/Bulle/apps/api/internal/hardware"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

const pollInterval = 3 * time.Second

func (h *Handler) pairStart(w http.ResponseWriter, r *http.Request) {
	var body ClaimRequest
	if err := httpjson.DecodeJSON(w, r, &body); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	device, err := newDevice(0, body)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	pairing, err := h.cfg.Pairings.Start(device.HardwareID, device.DeviceClass, device.Name)
	if errors.Is(err, internal.ErrPairingBusy) {
		httpjson.WriteError(w, troncerrors.Unavailable("too many pairings in progress, try again in a few minutes"))
		return
	}
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to start pairing", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusCreated, PairStartResponse{
		UserCode:   internal.FormatCode(pairing.UserCode),
		DeviceCode: pairing.DeviceCode,
		ExpiresIn:  int(internal.PairingTTL / time.Second),
		Interval:   int(pollInterval / time.Second),
	})
}

func (h *Handler) pairPoll(w http.ResponseWriter, r *http.Request) {
	var body PairPollRequest
	if err := httpjson.DecodeJSON(w, r, &body); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	grant, found, claimed := h.cfg.Pairings.Poll(body.DeviceCode)
	switch {
	case !found:
		httpjson.WriteError(w, troncerrors.NotFound("pairing expired or unknown, start again"))
	case !claimed:
		httpjson.WriteJSON(w, http.StatusOK, PairPollResponse{Status: "pending"})
	default:
		device := toResponse(grant.Device)
		httpjson.WriteJSON(w, http.StatusOK, PairPollResponse{
			Status:       "claimed",
			Device:       &device,
			DeviceToken:  grant.DeviceToken,
			GatewayWSURL: grant.GatewayWSURL,
		})
	}
}

func (h *Handler) claimCode(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	var body ClaimCodeRequest
	if err := httpjson.DecodeJSON(w, r, &body); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	pairing, ok := h.cfg.Pairings.Take(body.Code)
	if !ok {
		httpjson.WriteError(w, troncerrors.NotFound("no device is waiting with this code"))
		return
	}
	claimed, err := h.claimPairing(r, identity.UserID, pairing, body.Name)
	if err != nil {
		h.cfg.Pairings.Restore(pairing)
		h.claimFailed(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusCreated, toResponse(claimed))
}

func (h *Handler) claimPairing(r *http.Request, userID int64, pairing internal.Pairing, name string) (internal.Device, error) {
	device, err := newDevice(userID, ClaimRequest{
		HardwareID:  pairing.HardwareID,
		Name:        cmp.Or(strings.TrimSpace(name), pairing.Name),
		DeviceClass: pairing.DeviceClass,
	})
	if err != nil {
		return internal.Device{}, err
	}
	token, hash := internal.MintToken()
	claimed, err := h.cfg.Devices.Claim(r.Context(), device, hash)
	if err != nil {
		return internal.Device{}, err
	}
	h.cfg.Registry.Close(claimed.ID)
	h.cfg.Pairings.Complete(pairing, internal.Grant{Device: claimed, DeviceToken: token, GatewayWSURL: h.gatewayURL(r)})
	return claimed, nil
}

func (h *Handler) claimFailed(w http.ResponseWriter, err error) {
	if errors.Is(err, internal.ErrTooManyDevices) {
		httpjson.WriteError(w, troncerrors.Failed(fmt.Sprintf("an account pairs at most %d devices", internal.MaxDevicesPerUser)))
		return
	}
	var invalid *troncerrors.Error
	if errors.As(err, &invalid) {
		httpjson.WriteError(w, err)
		return
	}
	httpjson.WriteError(w, troncerrors.Internal("failed to claim the device", err))
}
