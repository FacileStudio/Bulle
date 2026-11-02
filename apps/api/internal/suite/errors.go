package suite

import (
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrNotConnected    = errors.New("Facile access needs to be reconnected: ask the user to sign out of Bulle and sign in again")
	ErrNoAccount       = errors.New("no account")
	ErrAppDown         = errors.New("app unavailable")
	ErrSubjectRejected = errors.New("suite: the user token was rejected")
	ErrWriteNotGranted = errors.New("write access not granted")
)

type AppError struct {
	App  string
	URL  string
	Kind error
}

func (e *AppError) Unwrap() error { return e.Kind }

func (e *AppError) Error() string {
	label := Label(e.App)
	if errors.Is(e.Kind, ErrWriteNotGranted) {
		return fmt.Sprintf("Bulle is not allowed to change %s data yet: the suite administrator must grant the %s.write scope in Registre.", label, e.App)
	}
	if errors.Is(e.Kind, ErrNoAccount) {
		return fmt.Sprintf("The user has no %s account yet. Ask them to open %s once, then try again.", label, e.URL)
	}
	return fmt.Sprintf("%s is not reachable right now. Try again later.", label)
}

func refusal(body []byte) string {
	var payload struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &payload) != nil || payload.Error.Message == "" {
		return ""
	}
	return ": " + payload.Error.Message
}
