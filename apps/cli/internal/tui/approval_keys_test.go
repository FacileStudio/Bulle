package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/FacileStudio/Bulle/apps/cli/internal/approval"
)

// Esc denies, and an upper-case answer counts the same as a lower-case one.
func TestApprovalTakesEscAndUpperCase(t *testing.T) {
	cases := []struct {
		press tea.KeyPressMsg
		want  approval.Decision
	}{
		{tea.KeyPressMsg{Code: tea.KeyEscape}, approval.Denied},
		{tea.KeyPressMsg{Code: 'y', Text: "Y", Mod: tea.ModShift}, approval.AllowedOnce},
		{tea.KeyPressMsg{Code: 'A', Text: "A"}, approval.AllowedForSession},
	}
	for _, tc := range cases {
		m := sized()
		decision := make(chan approval.Decision, 1)
		m.parkApproval(approval.Request{Name: "search", Decision: decision})
		m.Update(tc.press)
		select {
		case got := <-decision:
			if got != tc.want {
				t.Errorf("%q = %v, want %v", tc.press.String(), got, tc.want)
			}
		default:
			t.Errorf("%q answered nothing", tc.press.String())
		}
	}
}
