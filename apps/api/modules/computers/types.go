package computers

import "time"

type ComputerResponse struct {
	Status       string     `json:"status"`
	Error        string     `json:"error,omitempty"`
	CPUs         int        `json:"cpus"`
	MemoryMB     int        `json:"memory_mb"`
	DiskGB       int        `json:"disk_gb"`
	LastActiveAt *time.Time `json:"last_active_at,omitempty"`
}

type TagRequest struct {
	Tag string `json:"tag"`
}

type EmptyRequest struct{}
