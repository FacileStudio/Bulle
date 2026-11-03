package push

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

const streamHeartbeat = 15 * time.Second

// stream sends the user's notifications to an open desktop app as
// server-sent events, with a comment line every 15s to keep proxies open.
func (h *Handler) stream(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	notes, unsubscribe := h.hub.Subscribe(identity.UserID)
	defer unsubscribe()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("X-Accel-Buffering", "no")
	rc := http.NewResponseController(w)
	write := func(frame string) bool {
		_ = rc.SetWriteDeadline(time.Now().Add(time.Minute))
		if _, err := fmt.Fprint(w, frame); err != nil {
			return false
		}
		return rc.Flush() == nil
	}
	w.WriteHeader(http.StatusOK)
	if !write(": ok\n\n") {
		return
	}
	ticker := time.NewTicker(streamHeartbeat)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			if !write(": ping\n\n") {
				return
			}
		case n := <-notes:
			data, _ := json.Marshal(StreamNotification{Title: n.Title, Body: n.Body, Path: n.Path})
			if !write("event: notification\ndata: " + string(data) + "\n\n") {
				return
			}
		}
	}
}
