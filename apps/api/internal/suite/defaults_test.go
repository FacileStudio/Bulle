package suite

import (
	"testing"
)

func TestDefaultsKeepWhatTheModelNamed(t *testing.T) {
	d := Defaults{AgendaCalendar: 2, OpusWorkspace: "w"}
	if d.calendar(9) != 9 || d.workspace("x") != "x" {
		t.Fatal("a named value must win over the space default")
	}
	if d.calendar(0) != 2 || d.workspace("") != "w" {
		t.Fatal("an omitted value must take the space default")
	}
	if (Defaults{}).calendar(0) != 0 || (Defaults{}).workspace("") != "" {
		t.Fatal("an unset default must leave the app's own default")
	}
}
