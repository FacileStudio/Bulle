package memory

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// Backend stores memory files by slash-separated key, such as "MEMORY.md",
// "notes/2026-11-02.md" or "backups/MEMORY.md.2026-11-01". Read returns an
// error wrapping fs.ErrNotExist for a missing key. List returns the keys
// directly under a directory key, sorted.
type Backend interface {
	Read(key string) (string, error)
	Write(key string, data []byte) error
	List(dir string) ([]string, error)
}

// FileBackend keeps memory files under a directory on disk.
type FileBackend struct {
	Dir string
}

func (b FileBackend) file(key string) string {
	return filepath.Join(b.Dir, filepath.FromSlash(key))
}

// Read returns the file stored at key.
func (b FileBackend) Read(key string) (string, error) {
	data, err := os.ReadFile(b.file(key))
	return string(data), err
}

// Write replaces the file at key atomically, with mode 0600.
func (b FileBackend) Write(key string, data []byte) error {
	target := b.file(key)
	dir := filepath.Dir(target)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), target)
}

// List returns the regular files directly under dir as keys.
func (b FileBackend) List(dir string) ([]string, error) {
	entries, err := os.ReadDir(b.file(dir))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var keys []string
	for _, e := range entries {
		if e.Type().IsRegular() && !strings.HasPrefix(e.Name(), ".") {
			keys = append(keys, path.Join(dir, e.Name()))
		}
	}
	sort.Strings(keys)
	return keys, nil
}
