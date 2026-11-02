// Package memory is bulle's bounded, file-backed assistant memory: a small
// profile, a small long-term file, add-only daily notes and backups.
package memory

import (
	"errors"
	"io/fs"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/FacileStudio/Bulle/core/settings"
)

const (
	// UserFile is the profile file name.
	UserFile = "USER.md"
	// MemoryFile is the long-term memory file name.
	MemoryFile = "MEMORY.md"
	// UserCap is the character cap of USER.md.
	UserCap = 1500
	// MemoryCap is the character cap of MEMORY.md.
	MemoryCap = 2500
	// MaxFact is the longest fact Remember accepts.
	MaxFact = 500
)

var (
	// ErrEmptyFact is returned for a blank fact.
	ErrEmptyFact = errors.New("memory: empty fact")
	// ErrFactTooLong is returned for a fact over MaxFact characters.
	ErrFactTooLong = errors.New("memory: fact too long")
	// ErrOverCap is returned when a write would exceed a file's cap.
	ErrOverCap = errors.New("memory: write would exceed the file cap")
	// ErrVerify is returned when the read-back after a write does not match.
	ErrVerify = errors.New("memory: read-back verification failed")
)

// Store owns one user's memory, kept in a Backend.
type Store struct {
	files Backend
	mu    sync.Mutex
	clock func() time.Time
}

// New returns a Store keeping its files under dir on disk.
func New(dir string) *Store {
	return NewStore(FileBackend{Dir: dir})
}

// NewStore returns a Store over any Backend.
func NewStore(files Backend) *Store {
	return &Store{files: files, clock: time.Now}
}

// Default returns a Store rooted at memory/ inside settings.DataDir.
func Default() (*Store, error) {
	dir, err := settings.DataDir()
	if err != nil {
		return nil, err
	}
	return New(filepath.Join(dir, "memory")), nil
}

func (s *Store) path(parts ...string) string {
	return path.Join(parts...)
}

func (s *Store) readFile(key string) (string, error) {
	data, err := s.files.Read(key)
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	return data, err
}

func (s *Store) writeAtomic(key string, data []byte) error {
	return s.files.Write(key, data)
}

func (s *Store) backup(name, suffix, data string) error {
	return s.writeAtomic(s.path("backups", name+"."+suffix), []byte(data))
}
