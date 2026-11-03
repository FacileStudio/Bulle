package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintStatusReportsFailure(t *testing.T) {
	var buf bytes.Buffer
	lines := []statusLine{{statusOK, "a", "fine"}, {statusWarn, "b", "meh"}}
	if printStatus(&buf, lines) {
		t.Fatal("ok and warn lines must not fail the status")
	}
	if !printStatus(&buf, append(lines, statusLine{statusFail, "c", "broken"})) {
		t.Fatal("a FAIL line must fail the status")
	}
	if !strings.Contains(buf.String(), "FAIL  c") {
		t.Errorf("output = %q", buf.String())
	}
}
