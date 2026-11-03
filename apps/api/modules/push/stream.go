package push

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	pushsender "github.com/FacileStudio/Bulle/apps/api/internal/push"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

const streamHeartbeat = 15 * time.Second

// stream sends the user's live events (notifications, session, run and
// approval changes) to an open app as server-sent events. The stream ends
// when the listener falls behind; the client then reconnects and refetches
// what it shows.
func (h *Handler) stream(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	events, unsubscribe := h.hub.Subscribe(identity.UserID)
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
	if write(": ok\n\n") {
		pump(r.Context(), events, write)
	}
}

// pump writes each event as it comes, with a comment line every 15s to keep
// proxies open, until the client leaves, a write fails or the hub closes the
// channel.
func pump(ctx context.Context, events <-chan pushsender.Event, write func(string) bool) {
	ticker := time.NewTicker(streamHeartbeat)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !write(": ping\n\n") {
				return
			}
		case event, open := <-events:
			if !open || !write("event: "+event.Type+"\ndata: "+string(event.Data)+"\n\n") {
				return
			}
		}
	}
}
