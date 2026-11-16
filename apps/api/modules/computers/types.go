package computers

import "time"

type ComputerResponse struct {
	Status       string     `json:"status"`
	Error        string     `json:"error,omitempty"`
	CPUs         int        `json:"cpus"`
	MemoryMB     int        `json:"memory_mb"`
	DiskGB       int        `json:"disk_gb"`
	Network      string     `json:"network"`
	NetAllow     []string   `json:"net_allow,omitempty"`
	LastActiveAt *time.Time `json:"last_active_at,omitempty"`
}

type TagRequest struct {
	Tag string `json:"tag"`
}

// NetworkRequest sets the computer's egress mode: open, offline or allowlist.
// NetAllow lists the extra hosts of the allowlist mode (host, *.host or
// host:port, at most 64) and replaces the stored list; leaving it out clears
// it. It is refused with any other mode.
type NetworkRequest struct {
	Network  string   `json:"network"`
	NetAllow []string `json:"net_allow,omitempty"`
}

type EmptyRequest struct{}
