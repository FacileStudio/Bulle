package assistant

import (
	"errors"
)

// ErrLocked means another assistant run holds the run lock.
var ErrLocked = errors.New("assistant: another run holds the lock")

func withLock(path string, exclusive bool, fn func() error) error {
	unlock, err := lockFile(path, exclusive, true)
	if err != nil {
		return err
	}
	defer unlock()
	return fn()
}
