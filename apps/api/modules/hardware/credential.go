package hardware

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	internal "github.com/FacileStudio/Bulle/apps/api/internal/hardware"
)

var errRevoked = errors.New("hardware: credential revoked")

// credentialCheck answers nil while the socket's credential still holds,
// errRevoked once it is gone, and any other error when it cannot tell.
type credentialCheck func(ctx context.Context) error

func (h *Handler) tokenValid(ctx context.Context, hash, deviceID string) error {
	device, err := h.cfg.Sockets.ByTokenHash(ctx, hash)
	if errors.Is(err, internal.ErrNotFound) || (err == nil && device.ID != deviceID) {
		return errRevoked
	}
	return err
}

func (h *Handler) relayValid(ctx context.Context, r *http.Request, userID int64, deviceID string) error {
	if err := h.sessionValid(r.Clone(ctx), userID); err != nil {
		return err
	}
	_, err := h.cfg.Sockets.Owned(ctx, userID, deviceID)
	if errors.Is(err, internal.ErrNotFound) {
		return errRevoked
	}
	return err
}

func (h *Handler) sessionValid(r *http.Request, userID int64) error {
	recorder := &statusRecorder{header: http.Header{}, status: http.StatusOK}
	var current authcontext.Identity
	h.requireAuth(http.HandlerFunc(func(_ http.ResponseWriter, authed *http.Request) {
		current, _ = authcontext.IdentityFromContext(authed.Context())
	})).ServeHTTP(recorder, r)
	if current.UserID == userID {
		return nil
	}
	if recorder.status >= http.StatusInternalServerError {
		return fmt.Errorf("hardware: the session check answered %d", recorder.status)
	}
	return errRevoked
}

type statusRecorder struct {
	header http.Header
	status int
}

func (s *statusRecorder) Header() http.Header         { return s.header }
func (s *statusRecorder) Write(b []byte) (int, error) { return len(b), nil }
func (s *statusRecorder) WriteHeader(status int)      { s.status = status }
