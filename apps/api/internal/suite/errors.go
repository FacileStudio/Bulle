package suite

import (
	"errors"
	"fmt"
)

var (
	ErrNotConnected    = errors.New("Facile access needs to be reconnected: ask the user to sign out of Bulle and sign in again")
	ErrNoAccount       = errors.New("no account")
	ErrAppDown         = errors.New("app unavailable")
	ErrSubjectRejected = errors.New("suite: the user token was rejected")
)

type AppError struct {
	App  string
	URL  string
	Kind error
}

func (e *AppError) Unwrap() error { return e.Kind }

func (e *AppError) Error() string {
	label := Label(e.App)
	if errors.Is(e.Kind, ErrNoAccount) {
		return fmt.Sprintf("The user has no %s account yet. Ask them to open %s once, then try again.", label, e.URL)
	}
	return fmt.Sprintf("%s is not reachable right now. Try again later.", label)
}
