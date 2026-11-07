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
		if mark := fenceMark(line); mark != "" && (open == "" || open == mark) {
			open = toggle(open, mark)
		}
		offset += len(line) + 1
	}
	return last
}

// fenceMark is the fence a line opens or closes, or "" for any other line.
func fenceMark(line string) string {
	trimmed := strings.TrimLeft(line, " ")
	for _, mark := range []string{"```", "~~~"} {
		if strings.HasPrefix(trimmed, mark) {
			return mark
		}
	}
	return ""
}

func toggle(open, mark string) string {
	if open == mark {
		return ""
	}
	return mark
}

// openFence is the byte offset of the line that opens a fence text never
// closes, or -1 when every fence is closed.
func openFence(text string) int {
	at, offset, open := -1, 0, ""
	for line := range strings.SplitSeq(text, "\n") {
		if mark := fenceMark(line); mark != "" && (open == "" || open == mark) {
			open = toggle(open, mark)
			at = offset
		}
		offset += len(line) + 1
	}
	if open == "" {
		return -1
	}
	return at
}

// fenceTail cuts text inside a long open fence down to the fence's header and
// its last keep source lines. Everything above them is off the live region
// anyway, and rendering it on every delta is what made a long code block
// stream slower the longer it got.
func fenceTail(text string, keep int) string {
	at := openFence(text)
	if at < 0 {
		return text
	}
	header, body, found := strings.Cut(text[at:], "\n")
	if !found || strings.Count(body, "\n") < keep {
		return text
	}
	cut := len(body)
	for range keep {
		cut = strings.LastIndexByte(body[:cut], '\n')
	}
	return header + "\n" + body[cut+1:]
}
