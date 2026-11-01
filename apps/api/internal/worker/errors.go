package worker

import (
	stderrors "errors"
	"fmt"

	"github.com/FacileStudio/tronc/errors"
)

const (
	CodeUnauthorized = "unauthorized"
	CodeInvalid      = "invalid"
	CodeNotFound     = "not_found"
	CodeExists       = "exists"
	CodeNotStopped   = "not_stopped"
	CodeBusy         = "busy"
	CodeInternal     = "internal"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"error"`
	Status  int    `json:"-"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("boite worker %s: %s", e.Code, e.Message)
}

func IsCode(err error, code string) bool {
	var workerErr *Error
	return stderrors.As(err, &workerErr) && workerErr.Code == code
}

func envelope(err *Error) error {
	var out *errors.Error
	switch err.Code {
	case CodeInvalid:
		out = errors.Invalid(err.Message)
	case CodeNotFound:
		out = errors.NotFound("computer not found")
	case CodeExists:
		out = errors.Conflict("computer already exists")
	case CodeNotStopped:
		out = errors.Failed("the computer must be stopped first")
	case CodeBusy:
		out = errors.Unavailable("the computer is busy, try again shortly")
	case CodeUnauthorized:
		out = errors.Internal("the computer service refused our credentials", nil)
	default:
		out = errors.Internal("the computer service failed", nil)
	}
	out.Cause = err
	return out
}

func unavailable(cause error) error {
	return errors.New("unavailable", "the computer service is unreachable", cause)
}
