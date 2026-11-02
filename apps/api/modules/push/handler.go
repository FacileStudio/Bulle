package push

import (
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	pushsender "github.com/FacileStudio/Bulle/apps/api/internal/push"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

type Handler struct {
	sender *pushsender.Sender
}

func NewHandler(sender *pushsender.Sender) *Handler {
	return &Handler{sender: sender}
}

func (h *Handler) status(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	if !h.sender.Enabled() {
		httpjson.WriteJSON(w, http.StatusOK, StatusResponse{})
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, StatusResponse{
		Enabled:      true,
		Server:       h.sender.Server(),
		Topic:        h.sender.Topic(identity.UserID),
		SubscribeURL: h.sender.SubscribeURL(identity.UserID),
	})
}

func (h *Handler) test(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	if !h.sender.Enabled() {
		httpjson.WriteError(w, troncerrors.Conflict("push is not configured on this server"))
		return
	}
	h.sender.Notify(r.Context(), identity.UserID, pushsender.Notification{Title: "Bulle", Body: "Test notification", Path: ""})
	w.WriteHeader(http.StatusNoContent)
}
