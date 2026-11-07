package tui

import "testing"

func TestSafeBreakNeverSplitsAnOpenFence(t *testing.T) {
	cases := map[string]int{
		"para one\n\npara two":                        8,
		"intro\n\n```go\npackage main\n\nfunc main()": 5,
		"```go\na\n\nb":                               -1,
		"```\na\n```\n\nafter":                        9,
		"abc\n\n":                                     3,
	}
	for text, want := range cases {
		if got := safeBreak(text); got != want {
			t.Errorf("safeBreak(%q) = %d, want %d", text, got, want)
		}
	}
}
