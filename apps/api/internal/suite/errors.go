package suite

import (
	"errors"
	"fmt"
)

var (
	ErrNotConnected = errors.New("the Facile access needs to be reconnected: ask the user to sign out of Bulle and sign in again")
	ErrNoAccount    = errors.New("no account")
	ErrNoAccess     = errors.New("no access")
	ErrAppDown      = errors.New("app unavailable")
)

type AppError struct {
	App  string
	URL  string
	Kind error
}

func (e *AppError) Unwrap() error { return e.Kind }

func (e *AppError) Error() string {
	label := Label(e.App)
	if errors.Is(e.Kind, ErrNoAccess) {
		return fmt.Sprintf("The user may not use %s: only a suite administrator can give them access in Registre.", label)
	}
	if errors.Is(e.Kind, ErrNotConnected) {
		return ErrNotConnected.Error()
	}
	if errors.Is(e.Kind, ErrNoAccount) {
		return fmt.Sprintf("The user has no %s account yet. Ask them to open %s once, then try again.", label, e.URL)
	}
	return fmt.Sprintf("%s is not reachable right now. Try again later.", label)
}

// RefusedError is an app's answer outside 2xx that is neither a missing
// account nor an outage, such as a 404 or a 400. It keeps the status so a
// route can pass it on.
type RefusedError struct {
	App    string
	Status int
	Reason string
}

func (e *RefusedError) Error() string {
	return fmt.Sprintf("%s answered %d%s", Label(e.App), e.Status, e.Reason)
}
