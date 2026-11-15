// Package encretest installs a stand-in encre binary for tests, so the write
// paths can be exercised without the real tool on the machine.
package encretest

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/encre"
)

// LogEnv names the file the fake appends every input it receives to.
const LogEnv = "FAKE_ENCRE_LOG"

const separator = "\n---\n"

const script = `#!/bin/sh
input=$(cat)
if [ -n "${FAKE_ENCRE_LOG:-}" ]; then printf '%s\n---\n' "$input" >> "$FAKE_ENCRE_LOG"; fi
case "$input" in
  *INVALID*) printf '%s\n' '{"errors":[{"column":1,"line":3,"message":"text is empty or holds a forbidden character","path":"doc.content[1]"}],"valid":false}'; exit 1 ;;
  *OVERSIZE*) printf '%s\n' '{"error":{"code":"invalid","exit":1,"message":"input is larger than 8388608 bytes; send less or raise --max-bytes","path":null}}' >&2; exit 1 ;;
  *CRASH*) printf '%s\n' '{"error":{"code":"internal","exit":3,"message":"panic","path":null}}' >&2; exit 3 ;;
esac
printf '%s\n' '{"errors":[],"valid":true}'
`

// Install writes the fake and points ENCRE_BIN at it for the rest of the
// test. The fake refuses any input holding INVALID with a verdict, any input
// holding OVERSIZE with an input error, and exits 3 on CRASH. It answers the
// path of the log file every input is appended to.
func Install(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "encre")
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(dir, "inputs.log")
	t.Setenv(encre.BinaryEnv, bin)
	t.Setenv(LogEnv, log)
	return log
}

// Nested is a document whose custom fences open depth callout blocks on one
// path, each fence one backtick shorter than the one around it. encre reads
// up to eight; the ninth fence, on line depth + 2, is refused at its place.
func Nested(depth int) string {
	lines := []string{"# Deep", ""}
	for i := range depth {
		lines = append(lines, strings.Repeat("`", depth+3-i)+`callout tone="info" title="L`+strconv.Itoa(i+1)+`"`)
	}
	lines = append(lines, "deep")
	for i := depth - 1; i >= 0; i-- {
		lines = append(lines, strings.Repeat("`", depth+3-i))
	}
	return strings.Join(lines, "\n") + "\n"
}

// Inputs answers everything the fake has validated so far, one entry per call.
func Inputs(t *testing.T, log string) []string {
	t.Helper()
	raw, err := os.ReadFile(log)
	if err != nil {
		return nil
	}
	return strings.Split(strings.TrimSuffix(string(raw), separator), separator)
}
