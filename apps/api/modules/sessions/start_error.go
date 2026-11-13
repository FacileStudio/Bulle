package sessions

import (
	"errors"
	"fmt"

	"github.com/FacileStudio/Bulle/apps/api/internal/plans"
	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	troncerrors "github.com/FacileStudio/tronc/errors"
)

// startError is the HTTP answer to a run that could not start.
func startError(err error) error {
	switch {
	case errors.Is(err, runner.ErrBadPage):
		return troncerrors.Invalid("page must name a docs, toile or feuilles page with an id and an https url")
	case errors.Is(err, runner.ErrEmptyMessage):
		return troncerrors.Invalid("message cannot be empty")
	case errors.Is(err, runner.ErrTooManyAttachments):
		return troncerrors.Invalid(fmt.Sprintf("at most %d attachments per message", runner.MaxAttachments))
	case errors.Is(err, runner.ErrUnknownAttachment):
		return troncerrors.Invalid("unknown attachment")
	case errors.Is(err, runner.ErrTooManyRuns):
		return troncerrors.RateLimited("too many active runs")
	case errors.Is(err, runner.ErrShuttingDown):
		return troncerrors.Unavailable("server is shutting down")
	case errors.Is(err, plans.ErrAIAccess):
		return err
	case errors.Is(err, store.ErrNotFound):
		return troncerrors.NotFound("session not found")
	case errors.Is(err, store.ErrSessionBusy):
		return troncerrors.Conflict("session already has an active run")
	default:
		return troncerrors.Internal("failed to start run", err)
	}
}
