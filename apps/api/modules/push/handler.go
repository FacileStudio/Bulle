package push

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	pushsender "github.com/FacileStudio/Bulle/apps/api/internal/push"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

type Handler struct {
	hub *pushsender.Hub
}

func NewHandler(hub *pushsender.Hub) *Handler {
	return &Handler{hub: hub}
}

func (h *Handler) status(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	ntfy := h.hub.Ntfy()
	httpjson.WriteJSON(w, http.StatusOK, StatusResponse{
		Enabled:      ntfy.Enabled(),
		Server:       ntfy.Server(),
		Topic:        ntfy.Topic(identity.UserID),
		SubscribeURL: ntfy.SubscribeURL(identity.UserID),
		Native:       h.hub.Native(),
		WebPushKey:   h.hub.WebPushKey(),
	})
}

func (h *Handler) test(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	if !h.hub.Enabled() {
		httpjson.WriteError(w, troncerrors.Conflict("push is not configured on this server"))
		return
	}
	h.hub.Notify(r.Context(), identity.UserID, pushsender.Notification{Title: "Bulle", Body: "Test notification", Path: ""})
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) registerDevice(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	var body DeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Token) == "" {
		httpjson.WriteError(w, troncerrors.Invalid("a device token is required"))
		return
	}
	if body.Kind == pushsender.KindWeb && (body.P256dh == "" || body.Auth == "") {
		httpjson.WriteError(w, troncerrors.Invalid("a browser subscription needs its p256dh and auth keys"))
		return
	}
	err := h.hub.Register(r.Context(), pushsender.Device{
		UserID: identity.UserID, Kind: body.Kind, Token: strings.TrimSpace(body.Token), P256dh: body.P256dh, Auth: body.Auth,
	})
	if errors.Is(err, pushsender.ErrUnsupported) {
		httpjson.WriteError(w, troncerrors.Conflict("this server cannot send to that kind of device"))
		return
	}
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to save the device", err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) forgetDevice(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	var body ForgetDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Token == "" {
		httpjson.WriteError(w, troncerrors.Invalid("a device token is required"))
		return
	}
	if err := h.hub.Unregister(r.Context(), identity.UserID, body.Token); err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to forget the device", err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
