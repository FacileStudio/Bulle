// Package encre validates artifacts with the encre command-line tool, the
// reference implementation of the Facile document format, through encre's Go
// module. Artifacts are markdown in encre's lossless dialect, and encre is the
// one judge of whether a body is a valid document.
package encre

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	encre "github.com/FacileStudio/encre/go"
)

// BinaryEnv names the environment variable that points at the encre binary.
// Unset, the binary is looked up on PATH as "encre".
const BinaryEnv = "ENCRE_BIN"

// MaxBytes is the largest body ValidateMarkdown accepts. A crafted document
// costs encre about a second of CPU per megabyte, so every write path refuses
// a larger body before encre sees it.
const MaxBytes = 1 << 20

const timeout = 20 * time.Second

// ErrUnavailable is returned when no encre binary can be found.
var ErrUnavailable = errors.New("encre is not installed")

// ErrTooLarge is returned for a body over MaxBytes, before encre runs.
var ErrTooLarge = fmt.Errorf("the document is larger than %d MiB", MaxBytes>>20)

// Problem is one reason a document is invalid, as encre reports it. Line and
// Column are 1-based and name where the offending block starts, when known.
type Problem = encre.Problem

// ValidationError is encre's verdict on a body it refuses.
type ValidationError struct {
	Problems []Problem
}

func (e *ValidationError) Error() string {
	messages := make([]string, 0, len(e.Problems))
	for _, p := range e.Problems {
		messages = append(messages, describe(p))
	}
	return "invalid document: " + strings.Join(messages, "; ")
}

func describe(p Problem) string {
	if p.Line > 0 {
		return fmt.Sprintf("line %d, column %d: %s", p.Line, p.Column, p.Message)
	}
	return p.Message
}

// Available reports whether an encre binary can be found.
func Available() bool {
	_, err := exec.LookPath(encre.Binary())
	return err == nil
}

// ValidateMarkdown runs `encre validate --from markdown --json` on body. It
// answers nil for a valid document, ErrTooLarge for a body over MaxBytes, a
// *ValidationError for one encre refuses, ErrUnavailable when encre is not
// installed, and another error when encre could not run.
func ValidateMarkdown(ctx context.Context, body []byte) error {
	if len(body) > MaxBytes {
		return ErrTooLarge
	}
	if !Available() {
		return ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	verdict, err := encre.Validate(ctx, body, encre.Markdown)
	var refused *encre.Error
	switch {
	case errors.As(err, &refused) && refused.Invalid():
		return &ValidationError{Problems: []Problem{{Message: refused.Message}}}
	case err != nil:
		return fmt.Errorf("encre validate: %w", err)
	case verdict.Valid:
		return nil
	}
	return &ValidationError{Problems: verdict.Errors}
}
