package runner

import (
	"github.com/FacileStudio/nacelle"
)

// readMessage records the URLs one stored message shows the run. It answers
// the level the run has reached after it, given level before it, and how many
// links from results the message's calls followed. A message that opens a
// chat from an idea taints the run, and its text is not the owner's.
func (g *gate) readMessage(own, seen seenURLs, level taint, message nacelle.Message) (taint, int) {
	owner, followed := message.Role == nacelle.RoleUser, 0
	if ideaStarted(message) {
		level, owner = tainted, false
	}
	for _, part := range message.Parts {
		if call, ok := part.(nacelle.ToolCall); ok && level != clean {
			followed += followedLink(own, seen, call)
		}
		level = max(level, g.readPart(own, seen, owner, part))
	}
	return level, followed
}

// readPart records the URLs one stored part shows the run, and reports how
// far it brought untrusted text into the context. The caller holds mu.
func (g *gate) readPart(own, seen seenURLs, owner bool, part nacelle.Part) taint {
	switch part := part.(type) {
	case nacelle.ToolCall:
		if g.sourceLocked(part.Name) {
			return tainted
		}
	case nacelle.ToolResult:
		level := g.taintOf(part.Name)
		if level != clean {
			seen.addResult(part.Result)
		}
		return level
	case nacelle.Text:
		if _, attached := attachmentOf(part.Text); attached {
			return tainted
		}
		if owner {
			own.addOwn(part.Text)
		}
	}
	return clean
}
