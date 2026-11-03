package matrixbot

import "testing"

func TestClear(t *testing.T) {
	const owner, agent, colleague = "@yann:p", "@pixel:p", "@marc:p"
	cases := []struct {
		name    string
		sender  string
		owner   string
		members []string
		want    clearance
	}{
		{"owner in a direct message", owner, owner, []string{owner, agent}, clearance{Owner: true, Private: true, Direct: true}},
		{"owner in a team room", owner, owner, []string{owner, agent, colleague}, clearance{Owner: true}},
		{"colleague in a team room", colleague, owner, []string{owner, agent, colleague}, clearance{}},
		{"colleague in a direct message", colleague, owner, []string{colleague, agent}, clearance{Direct: true}},
		{"owner unknown", owner, "", []string{owner, agent}, clearance{Direct: true}},
		{"owner id in another case", "@Yann:p", owner, []string{owner, agent}, clearance{Owner: true, Private: true, Direct: true}},
	}
	for _, c := range cases {
		if got := assess(c.sender, c.owner, agent, c.members); got != c.want {
			t.Errorf("%s: assess = %+v, want %+v", c.name, got, c.want)
		}
	}
}
