package memory

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

// routeError keeps an answer the handler already shaped, and turns a Mycelium
// failure into one: a refusal keeps its meaning, and anything else stays
// opaque.
func routeError(err error) error {
	var known *troncerrors.Error
	if errors.As(err, &known) {
		return known
	}
	var refused *suite.RefusedError
	switch {
	case disconnected(err):
		return troncerrors.Failed("mycelium is not connected")
	case errors.Is(err, suite.ErrAppDown):
		return troncerrors.Unavailable("mycelium is not reachable right now")
	case errors.Is(err, suite.ErrWriteNotGranted):
		return troncerrors.Forbidden("bulle may not change mycelium data")
	case !errors.As(err, &refused):
		return troncerrors.Internal("failed to reach mycelium", err)
	case refused.Status == http.StatusNotFound:
		return troncerrors.NotFound("memory page not found")
	case refused.Status == http.StatusForbidden:
		return troncerrors.Forbidden("this mycelium space is not yours")
	}
	return troncerrors.Invalid("mycelium refused the request" + refused.Reason)
}
