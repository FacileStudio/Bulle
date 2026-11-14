package runner

import (
	"slices"
	"strings"
)

const (
	writeScope  = "this tool in a tainted run"
	typingScope = "typing on "
)

// destructiveVerbs name a suite write that loses or moves data, so a tainted
// run asks before each one even after the owner approved the tool.
var destructiveVerbs = []string{"delete", "trash", "move", "clear", "revoke", "replace"}

// taintScope is how far the owner may approve a card the taint raised: later
// calls of the same suite write, or typing on the same page origin, which the
// browser policy gives as scope. The shell, memory writes, typing on a page
// whose origin is unknown, and a write that deletes, trashes, moves, clears,
// revokes or replaces stay asked each time.
func (g *gate) taintScope(name, origin string) string {
	switch {
	case name == browserTypeTool && origin != "":
		return typingScope + origin
	case name == shellTool, name == memoryWriteTool, name == browserTypeTool,
		!g.takesGrant(name), hasVerb(name, destructiveVerbs):
		return ""
	}
	return writeScope
}

func taintGrant(scope string) bool {
	return scope == writeScope || strings.HasPrefix(scope, typingScope)
}

// hasVerb reports whether a suite tool's name, past its app prefix, holds one
// of verbs.
func hasVerb(name string, verbs []string) bool {
	parts := strings.Split(name, "_")[1:]
	return slices.ContainsFunc(parts, func(part string) bool { return slices.Contains(verbs, part) })
}
