package agent

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestWithCommitCoAuthorAppendsTheTrailer(t *testing.T) {
	suffix := ` -m "` + coAuthorTrailer + `"`
	cases := map[string]string{
		`git commit -m "add feature"`:                                   `git commit -m "add feature"` + suffix,
		`git commit -m 'add feature'`:                                   `git commit -m 'add feature'` + suffix,
		`git commit --message "add feature"`:                            `git commit --message "add feature"` + suffix,
		`git commit -am "add feature"`:                                  `git commit -am "add feature"` + suffix,
		`git commit -m add`:                                             `git commit -m add` + suffix,
		`git commit -m "add feature" && git push`:                       `git commit -m "add feature"` + suffix + ` && git push`,
		`git add . && git commit -m "msg"`:                              `git add . && git commit -m "msg"` + suffix,
		`git commit -m "title" -m "body"`:                               `git commit -m "title" -m "body"` + suffix,
		`git -C . commit -m "msg"`:                                      `git -C . commit -m "msg"` + suffix,
		`git commit -m "already there` + "\n\n" + coAuthorTrailer + `"`: `git commit -m "already there` + "\n\n" + coAuthorTrailer + `"`,
	}
	for input, want := range cases {
		if got := withCommitCoAuthor(input, true); got != want {
			t.Errorf("withCommitCoAuthor(%q)\n got %q\nwant %q", input, got, want)
		}
	}
}

func TestWithCommitCoAuthorLeavesOtherCommandsAlone(t *testing.T) {
	for _, input := range []string{
		`echo "git commit is fun"`,
		`git log --oneline`,
		`git commit`,
		`git commit --amend --no-edit`,
		`git status`,
	} {
		if got := withCommitCoAuthor(input, true); got != input {
			t.Errorf("withCommitCoAuthor(%q) = %q, want it unchanged", input, got)
		}
	}
}

func TestWithCommitCoAuthorRespectsOptOut(t *testing.T) {
	input := `git commit -m "add feature"`
	if got := withCommitCoAuthor(input, false); got != input {
		t.Errorf("withCommitCoAuthor(%q, false) = %q, want it unchanged", input, got)
	}
}

func TestCoAuthorInputRewritesTheCommandJSON(t *testing.T) {
	in := json.RawMessage(`{"command":"git commit -m \"add feature\""}`)
	var decoded shellCommandInput
	if err := json.Unmarshal(coAuthorInput(in, true), &decoded); err != nil {
		t.Fatalf("coAuthorInput: %v", err)
	}
	if !strings.Contains(decoded.Command, coAuthorTrailer) {
		t.Errorf("command = %q, want the trailer inside it", decoded.Command)
	}
	if got := string(coAuthorInput(in, false)); got != string(in) {
		t.Errorf("coAuthorInput(..., false) = %q, want it unchanged", got)
	}
}
