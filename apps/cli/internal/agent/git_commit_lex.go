package agent

import "strings"

const wordBreaks = " \t\r;&|()\n`"

type shellWord struct {
	start, end int
	sep        bool
}

// lexWords splits a command into words with their byte positions, treating
// quoted runs as one word and every unquoted shell separator ( ; & | ( ) `,
// newline ) as a word of its own. A # at a word start opens a comment to end
// of line.
func lexWords(command string) []shellWord {
	var words []shellWord
	for i := 0; i < len(command); {
		c := command[i]
		switch {
		case c == ' ' || c == '\t' || c == '\r':
			i++
		case strings.IndexByte(wordBreaks, c) >= 0:
			words = append(words, shellWord{start: i, end: i + 1, sep: true})
			i++
		case c == '#':
			for i < len(command) && command[i] != '\n' {
				i++
			}
		default:
			w, next := readShellWord(command, i)
			words = append(words, w)
			i = next
		}
	}
	return words
}

// readShellWord reads one word starting at i: quotes and backslashes are
// consumed inside it, and the word ends at unquoted whitespace or separator.
func readShellWord(command string, i int) (shellWord, int) {
	start := i
	for i < len(command) {
		c := command[i]
		if c == '\'' || c == '"' {
			i = quotedWordEnd(command, i)
		} else if c == '\\' {
			i = min(i+2, len(command))
		} else if strings.IndexByte(wordBreaks, c) >= 0 {
			break
		} else {
			i++
		}
	}
	return shellWord{start: start, end: i}, i
}

// isEnvAssignment reports whether word has the NAME=value shape a shell
// accepts before a command.
func isEnvAssignment(word string) bool {
	eq := strings.IndexByte(word, '=')
	return eq > 0 && word[0] != '-' && !strings.ContainsAny(word[:eq], "\"'")
}

// quotedWordEnd returns the position after the closing quote of the string
// that starts at i; an unterminated quote runs to the end of the command.
func quotedWordEnd(command string, i int) int {
	quote := command[i]
	for i++; i < len(command); i++ {
		if quote == '"' && command[i] == '\\' && i+1 < len(command) {
			i++
			continue
		}
		if command[i] == quote {
			return i + 1
		}
	}
	return len(command)
}
