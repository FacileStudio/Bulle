package assistant

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/FacileStudio/Bulle/core/settings"
)

// State keeps one user's assistant state as named JSON documents, such as
// "ledger", "watch" or "quiet". Load returns nil for a missing document.
// Update runs fn on the current bytes under an exclusive lock and stores what
// it returns; a nil result leaves the document unchanged. TryLock takes a
// named non-blocking lock and returns ErrLocked when another holder has it.
type State interface {
	Load(key string) ([]byte, error)
	Update(key string, fn func([]byte) ([]byte, error)) error
	TryLock(name string) (func(), error)
}

// FileState keeps each document as <key>.json in Dir, guarded by <key>.lock.
type FileState struct {
	Dir string
}

// DefaultState is a FileState in settings.StateDir.
func DefaultState() (FileState, error) {
	dir, err := settings.StateDir()
	return FileState{Dir: dir}, err
}

func (f FileState) file(name, ext string) string {
	return filepath.Join(f.Dir, name+ext)
}

// Load reads <key>.json under a shared lock.
func (f FileState) Load(key string) ([]byte, error) {
	var data []byte
	err := withLock(f.file(key, ".lock"), false, func() error {
		var err error
		data, err = f.read(key)
		return err
	})
	return data, err
}

// Update rewrites <key>.json atomically under an exclusive lock.
func (f FileState) Update(key string, fn func([]byte) ([]byte, error)) error {
	return withLock(f.file(key, ".lock"), true, func() error {
		current, err := f.read(key)
		if err != nil {
			return err
		}
		next, err := fn(current)
		if err != nil || next == nil {
			return err
		}
		return writeAtomic(f.file(key, ".json"), next)
	})
}

// TryLock takes <name>.lock without waiting.
func (f FileState) TryLock(name string) (func(), error) {
	return lockFile(f.file(name, ".lock"), true, false)
}

func (f FileState) read(key string) ([]byte, error) {
	data, err := os.ReadFile(f.file(key, ".json"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	return data, err
}
