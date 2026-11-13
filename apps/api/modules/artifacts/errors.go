package artifacts

import (
	"errors"
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	troncerrors "github.com/FacileStudio/tronc/errors"
)

var errNotConnected = errors.New("mycelium is not connected")

func disconnected(err error) bool {
	return errors.Is(err, errNotConnected) || errors.Is(err, suite.ErrNoAccount) || errors.Is(err, suite.ErrNotConnected)
}

// appError turns a Mycelium failure into the answer the client gets: a
// refusal keeps its meaning, and anything else stays opaque.
func appError(err error) error {
	var refused *suite.RefusedError
	switch {
	case disconnected(err):
		return troncerrors.Failed("mycelium is not connected")
	case errors.Is(err, suite.ErrAppDown):
		return troncerrors.Unavailable("mycelium is not reachable right now")
	case errors.Is(err, suite.ErrArtifactID):
		return troncerrors.Invalid("invalid artifact id")
	case !errors.As(err, &refused):
		return troncerrors.Internal("failed to reach mycelium", err)
	case refused.Status == http.StatusNotFound:
		return troncerrors.NotFound("artifact not found")
	case refused.Status == http.StatusForbidden:
		return troncerrors.Forbidden("this mycelium space is not yours")
	}
	return troncerrors.Invalid("mycelium refused the request" + refused.Reason)
}
