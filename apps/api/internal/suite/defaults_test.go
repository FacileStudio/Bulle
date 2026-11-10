package suite

import (
	"testing"
)

func TestDefaultsKeepWhatTheModelNamed(t *testing.T) {
	d := Defaults{OpusWorkspace: "w"}
	if d.workspace("x") != "x" {
		t.Fatal("a named value must win over the space default")
	}
	if d.workspace("") != "w" {
		t.Fatal("an omitted value must take the space default")
	}
	if (Defaults{}).workspace("") != "" {
		t.Fatal("an unset default must leave the app's own default")
	}
}
