package matrixbot

import "strings"

// clearance is what one inbound message is allowed to make the agent do, worked
// out from facts the homeserver vouches for and never from the message text.
//
// Owner is true only when the sender is the person the agent works for. Only
// then does the run get the owner's tools, memory and mail; anyone else talks
// to a guest session with no tools at all, so no wording of a request can reach
// the owner's data. Private is true when nobody but the owner and the agent can
// read the room; any other reply is screened for private data before it is
// posted. Direct is a one-to-one room, where the agent answers without being
// mentioned.
type clearance struct {
	Owner   bool
	Private bool
	Direct  bool
}

// assess decides the clearance of a message from sender in a room whose readers
// are members. An agent that does not know its owner yet treats every sender as
// a guest, so a missing owner fails closed.
func assess(sender, owner, self string, members []string) clearance {
	isOwner := owner != "" && same(sender, owner)
	private := isOwner
	for _, member := range members {
		if !same(member, owner) && !same(member, self) {
			private = false
		}
	}
	return clearance{Owner: isOwner, Private: private, Direct: len(members) == 2}
}

func same(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}
