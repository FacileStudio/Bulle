package suite

import (
	"testing"
)

func TestSpaceDefaultFillsAnOmittedAccount(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	tools, _, err := Tools(NewGateway(w.tokens(), w.cfg, nil), 1, Defaults{CourrierAccount: 42})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runTool(t, tools, "courrier_send", `{"to":["a@x"],"subject":"s","body":"b"}`); err != nil {
		t.Fatal(err)
	}
	if _, err := runTool(t, tools, "courrier_send", `{"account":5,"to":["a@x"],"subject":"s","body":"b"}`); err != nil {
		t.Fatal(err)
	}
	if len(w.writes) != 2 || w.writes[0].path != "/api/accounts/42/mail/send" || w.writes[1].path != "/api/accounts/5/mail/send" {
		t.Fatalf("want the space default when omitted and the named account otherwise, got %+v", w.writes)
	}
}

func TestDefaultsKeepWhatTheModelNamed(t *testing.T) {
	d := Defaults{CourrierAccount: 1, AgendaCalendar: 2, OpusWorkspace: "w"}
	if d.account(9) != 9 || d.calendar(9) != 9 || d.workspace("x") != "x" {
		t.Fatal("a named value must win over the space default")
	}
	if d.account(0) != 1 || d.calendar(0) != 2 || d.workspace("") != "w" {
		t.Fatal("an omitted value must take the space default")
	}
	if (Defaults{}).account(0) != 0 || (Defaults{}).workspace("") != "" {
		t.Fatal("an unset default must leave the app's own default")
	}
}
