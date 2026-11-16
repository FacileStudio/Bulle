package runner

import "strings"

// The egress modes of the user's computer, as the computers package names them.
const (
	NetworkOpen      = "open"
	NetworkOffline   = "offline"
	NetworkAllowlist = "allowlist"
)

// networkNote tells the model what its computer's network can reach, so it
// does not spend turns on an ssh remote or a blocked site. An open computer
// needs no note.
func networkNote(mode string, allow []string) string {
	switch mode {
	case NetworkOffline:
		return "Your computer has no network. Installs, downloads, git clones and web requests from its shell all fail. " +
			"Work with the files already there, and tell the user when a task needs the network."
	case NetworkAllowlist:
		return "Your computer's network is limited to a list of hosts. It reaches package registries (Debian mirrors, npm, PyPI, crates.io, the Go proxy) and GitHub over https" +
			extraHosts(allow) + ". " +
			"ssh, git@ and git:// remotes, ping, and every other site fail, so clone with https:// URLs. " +
			"When a host is blocked, do not retry or look for a way around it: tell the user which host to add in the computer's network settings."
	}
	return ""
}

func extraHosts(allow []string) string {
	if len(allow) == 0 {
		return ""
	}
	return ", plus the hosts the user added: " + strings.Join(allow, ", ")
}
