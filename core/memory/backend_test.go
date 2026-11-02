package memory

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
	"testing"
	"time"
)

type mapBackend map[string]string

func (m mapBackend) Read(key string) (string, error) {
	data, ok := m[key]
	if !ok {
		return "", fmt.Errorf("%s: %w", key, fs.ErrNotExist)
	}
	return data, nil
}

func (m mapBackend) Write(key string, data []byte) error {
	m[key] = string(data)
	return nil
}

func (m mapBackend) List(dir string) ([]string, error) {
	var keys []string
	for k := range m {
		if path.Dir(k) == dir {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys, nil
}

func TestStoreWorksOverAnyBackend(t *testing.T) {
	files := mapBackend{}
	s := NewStore(files)
	s.clock = func() time.Time { return time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC) }
	if _, err := s.Remember("Drinks green tea", s.clock()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(files["notes/2026-11-02.md"], "Drinks green tea") {
		t.Fatalf("note not written: %v", files)
	}
	res, err := s.ForgetReport("green tea")
	if err != nil || res.Removed != 1 {
		t.Fatalf("forget = %+v, %v", res, err)
	}
	if err := s.Restore(res.Backups[0]); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(files["notes/2026-11-02.md"], "Drinks green tea") {
		t.Fatalf("restore did not bring the note back: %v", files)
	}
}
