package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintStatusReportsFailure(t *testing.T) {
	var buf bytes.Buffer
	lines := []statusLine{{statusOK, "a", "fine"}, {statusWarn, "b", "meh"}}
	if failed, err := printStatus(&buf, lines); err != nil || failed {
		t.Fatalf("ok and warn lines must not fail the status: failed=%v err=%v", failed, err)
	}
	if failed, err := printStatus(&buf, append(lines, statusLine{statusFail, "c", "broken"})); err != nil || !failed {
		t.Fatalf("a FAIL line must fail the status: failed=%v err=%v", failed, err)
	}
	if !strings.Contains(buf.String(), "FAIL  c") {
		t.Errorf("output = %q", buf.String())
	}
}
