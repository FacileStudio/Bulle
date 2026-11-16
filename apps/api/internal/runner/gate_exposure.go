package runner

// privateReads are the tools that hand the run the owner's private data: the
// long-term memory and the past chats. A run that read one and was then
// tainted by the shell or a file holds both secrets and text a stranger may
// have written, which is what an open network lets out.
var privateReads = []string{"memory_read", "memory_search", ChatsSearchName}

// setNetwork records the egress mode of the computer the run uses. A mode the
// gate never learned reads as open.
func (g *gate) setNetwork(mode string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.trust.network = mode
}

// exposedLocked reports whether a soft-tainted run could carry private data
// out through the shell: its computer reaches the whole network and the run
// read memory or past chats. The caller holds mu.
func (g *gate) exposedLocked() bool {
	return g.trust.private && g.trust.network != NetworkOffline && g.trust.network != NetworkAllowlist
}

// softNeedLocked is whether a call waits for the owner in a run that only the
// shell or a file tainted: the shell does once the run could carry private data
// out. The caller holds mu.
func (g *gate) softNeedLocked(name string) ask {
	if name == shellTool && g.exposedLocked() {
		return askTaint
	}
	return askNo
}
