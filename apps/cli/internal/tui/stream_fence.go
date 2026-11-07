package tui

import "strings"

// safeBreak is the last blank-line break in text that sits outside every
// fenced code block, or -1 when there is none. A fragment committed inside an
// open fence would render as two broken blocks: the first unterminated, the
// second as plain text. It runs on every streamed delta, so it walks the text
// once, line by line, tracking the fence it is in.
func safeBreak(text string) int {
	last, offset, open := -1, 0, ""
	for line := range strings.SplitSeq(text, "\n") {
		if line == "" && offset > 0 && offset < len(text) && open == "" {
			last = offset - 1
		}
		trimmed := strings.TrimLeft(line, " ")
		for _, mark := range []string{"```", "~~~"} {
			if strings.HasPrefix(trimmed, mark) && (open == "" || open == mark) {
				open = toggle(open, mark)
			}
		}
		offset += len(line) + 1
	}
	return last
}

func toggle(open, mark string) string {
	if open == mark {
		return ""
	}
	return mark
}
