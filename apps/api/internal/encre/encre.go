// Package encre runs the encre command-line tool, the reference
// implementation of the Facile document format, as a subprocess. Artifacts
// are markdown in encre's lossless dialect, and encre is the one judge of
// whether a body is a valid document.
package encre

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// BinaryEnv names the environment variable that points at the encre binary.
// Unset, the binary is looked up on PATH as "encre".
const BinaryEnv = "ENCRE_BIN"

const timeout = 20 * time.Second

// ErrUnavailable is returned when no encre binary can be found.
var ErrUnavailable = errors.New("encre is not installed")

// Problem is one reason a document is invalid, as encre reports it.
type Problem struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

// ValidationError is encre's verdict on a body it refuses.
type ValidationError struct {
	Problems []Problem
}

func (e *ValidationError) Error() string {
	messages := make([]string, 0, len(e.Problems))
	for _, p := range e.Problems {
		messages = append(messages, p.Message)
	}
	return "invalid document: " + strings.Join(messages, "; ")
}

// Binary answers the path of the encre executable, or ErrUnavailable.
func Binary() (string, error) {
	path, err := exec.LookPath(cmp.Or(os.Getenv(BinaryEnv), "encre"))
	if err != nil {
		return "", ErrUnavailable
	}
	return path, nil
}

// Available reports whether an encre binary can be found.
func Available() bool {
	_, err := Binary()
	return err == nil
}

type verdict struct {
	Valid  bool      `json:"valid"`
	Errors []Problem `json:"errors"`
}

type failure struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// ValidateMarkdown runs `encre validate --from markdown --json` on body. It
// answers nil for a valid document, a *ValidationError for one encre refuses,
// ErrUnavailable when encre is not installed, and another error when encre
// could not run.
func ValidateMarkdown(ctx context.Context, body []byte) error {
	bin, err := Binary()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "validate", "--from", "markdown", "--json")
	cmd.Stdin = bytes.NewReader(body)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	runErr := cmd.Run()
	if ctx.Err() != nil {
		return fmt.Errorf("encre validate: %w", ctx.Err())
	}
	if runErr == nil {
		return nil
	}
	var exit *exec.ExitError
	if !errors.As(runErr, &exit) {
		return fmt.Errorf("encre validate: %w", runErr)
	}
	return refusal(exit.ExitCode(), stdout.Bytes(), stderr.Bytes())
}

// refusal reads a non-zero exit. An invalid document is a verdict on stdout;
// an input encre cannot read (not UTF-8, too large) is an error on stderr.
func refusal(code int, stdout, stderr []byte) error {
	var v verdict
	if json.Unmarshal(stdout, &v) == nil && !v.Valid && len(v.Errors) > 0 {
		return &ValidationError{Problems: v.Errors}
	}
	var f failure
	if json.Unmarshal(stderr, &f) == nil && f.Error.Message != "" {
		if f.Error.Code == "invalid" {
			return &ValidationError{Problems: []Problem{{Message: f.Error.Message}}}
		}
		return fmt.Errorf("encre validate: exit %d: %s", code, f.Error.Message)
	}
	return fmt.Errorf("encre validate: exit %d: %s", code, strings.TrimSpace(string(stderr)))
}
