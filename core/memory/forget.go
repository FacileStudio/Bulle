package memory

import (
	"errors"
	"path"
	"strings"
	"unicode/utf8"
)

const minForgetQuery = 4

var (
	// ErrEmptyQuery is returned by Forget for a blank query.
	ErrEmptyQuery = errors.New("memory: empty forget query")
	// ErrShortQuery is returned by Forget for a query under 4 characters.
	ErrShortQuery = errors.New("memory: forget query must be at least 4 characters")
)

// ForgetResult is what a Forget run removed and where the originals went.
// Each backup is backups/<file name>.<timestamp>; Restore puts one back.
type ForgetResult struct {
	Removed int
	Backups []string
}

// Forget removes every line containing query (case-insensitive) from USER.md,
// MEMORY.md and all notes, after backing each touched file up to
// backups/<name>.<timestamp>. It returns the number of lines removed.
func (s *Store) Forget(query string) (int, error) {
	res, err := s.ForgetReport(query)
	return res.Removed, err
}

// ForgetReport is Forget with the names of the backups it wrote. Blank lines
// and every untouched byte of a rewritten file stay as they were.
func (s *Store) ForgetReport(query string) (ForgetResult, error) {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return ForgetResult{}, ErrEmptyQuery
	}
	if utf8.RuneCountInString(q) < minForgetQuery {
		return ForgetResult{}, ErrShortQuery
	}
	notes, err := s.notes()
	if err != nil {
		return ForgetResult{}, err
	}
	paths := append([]string{s.path(UserFile), s.path(MemoryFile)}, notes...)
	stamp := s.clock().Format("20060102-150405.000")
	s.mu.Lock()
	defer s.mu.Unlock()
	var res ForgetResult
	for _, p := range paths {
		n, err := s.forgetIn(p, q, stamp)
		res.Removed += n
		if n > 0 {
			res.Backups = append(res.Backups, path.Base(p)+"."+stamp)
		}
		if err != nil {
			return res, err
		}
	}
	return res, nil
}

func (s *Store) forgetIn(key, q, stamp string) (int, error) {
	data, err := s.readFile(key)
	if err != nil || data == "" {
		return 0, err
	}
	var kept strings.Builder
	removed := 0
	for _, line := range strings.SplitAfter(data, "\n") {
		if strings.Contains(strings.ToLower(line), q) {
			removed++
			continue
		}
		kept.WriteString(line)
	}
	if removed == 0 {
		return 0, nil
	}
	if err := s.backup(path.Base(key), stamp, data); err != nil {
		return 0, err
	}
	return removed, s.writeAtomic(key, []byte(kept.String()))
}

func (s *Store) notes() ([]string, error) {
	keys, err := s.files.List("notes")
	if err != nil {
		return nil, err
	}
	var notes []string
	for _, k := range keys {
		if strings.HasSuffix(k, ".md") {
			notes = append(notes, k)
		}
	}
	return notes, nil
}
