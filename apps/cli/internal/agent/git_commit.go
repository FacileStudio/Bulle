package agent

import (
	"slices"
	"strings"
)

const coAuthorTrailer = "Co-authored-by: Bulle Agent <bulle@facile.studio>"

const coAuthorMessage = " -m \"" + coAuthorTrailer + "\""

var gitGlobalArgOpts = []string{"-C", "-c", "--git-dir", "--work-tree", "--namespace", "--exec-path", "--config-env", "--attr-source"}

type optArg int

const (
	optNone optArg = iota
	optNext
	optAttached
)

// withCommitCoAuthor appends the agent's co-author trailer to every git
// commit in command as a final -m paragraph, which git joins to the message
// with a blank line. It is idempotent: a command already carrying the trailer
// is returned untouched, and enabled=false turns the whole rewrite off.
func withCommitCoAuthor(command string, enabled bool) string {
	if !enabled || strings.Contains(command, coAuthorTrailer) {
		return command
	}
	inserts := commitInsertions(command)
	for at := len(inserts) - 1; at >= 0; at-- {
		i := inserts[at]
		command = command[:i] + coAuthorMessage + command[i:]
	}
	return command
}

// commitInsertions returns the byte positions, one per git commit command,
// where a trailing message paragraph begins. Each command is the run of words
// between unquoted shell separators; inserting last to first keeps every
// earlier position valid.
func commitInsertions(command string) []int {
	words := lexWords(command)
	var inserts []int
	for i := 0; i < len(words); {
		if words[i].sep {
			i++
			continue
		}
		end := i
		for end < len(words) && !words[end].sep {
			end++
		}
		if at := gitCommitMessageEnd(command, words[i:end]); at > 0 {
			inserts = append(inserts, at)
		}
		i = end
	}
	return inserts
}

// gitCommitMessageEnd returns the byte position after the last message
// argument of the words' git commit command, or -1 when this is not one or it
// carries no inline message for a trailer to join.
func gitCommitMessageEnd(command string, words []shellWord) int {
	k := skipCommandPrefix(command, words)
	if wordAt(command, words, k) != "git" {
		return -1
	}
	for k++; k < len(words); k++ {
		t := wordAt(command, words, k)
		if t == "commit" {
			return messageEnd(command, words[k+1:])
		}
		if !strings.HasPrefix(t, "-") {
			return -1
		}
		if skipsGlobalArg(command, words, k) {
			k++
		}
	}
	return -1
}

// skipCommandPrefix steps over the VAR=value words and env invocations a
// shell allows before the command itself.
func skipCommandPrefix(command string, words []shellWord) int {
	k := 0
	for k < len(words) {
		t := wordAt(command, words, k)
		if t != "env" && !isEnvAssignment(t) {
			break
		}
		k++
	}
	return k
}

// messageEnd returns the byte position after the last -m/--message argument,
// or -1 when the command has none. A bare -- ends the options.
func messageEnd(command string, words []shellWord) int {
	end := -1
	for k := 0; k < len(words); k++ {
		t := wordAt(command, words, k)
		if t == "--" {
			return end
		}
		switch messageOption(t) {
		case optNext:
			if k+1 < len(words) {
				end = words[k+1].end
				k++
			}
		case optAttached:
			end = words[k].end
		}
	}
	return end
}

// messageOption says how a word takes its message: as the following word (-m,
// --message, a flag cluster ending in m), inside the word itself (--message=x,
// -mmsg), or not at all.
func messageOption(t string) optArg {
	switch {
	case t == "-m" || t == "--message":
		return optNext
	case strings.HasPrefix(t, "--message="):
		return optAttached
	case strings.HasPrefix(t, "-") && !strings.HasPrefix(t, "--"):
		m := strings.IndexByte(t[1:], 'm')
		if m < 0 {
			return optNone
		}
		if m == len(t)-2 {
			return optNext
		}
		return optAttached
	}
	return optNone
}

// skipsGlobalArg reports whether the word at k is a git global option whose
// argument is the word after it, so the scan does not mistake that argument
// for the subcommand.
func skipsGlobalArg(command string, words []shellWord, k int) bool {
	if k+1 >= len(words) {
		return false
	}
	t := wordAt(command, words, k)
	return !strings.Contains(t, "=") && slices.Contains(gitGlobalArgOpts, t)
}

// wordAt is the word's text in command, or "" past the end.
func wordAt(command string, words []shellWord, k int) string {
	if k < 0 || k >= len(words) {
		return ""
	}
	return command[words[k].start:words[k].end]
}
