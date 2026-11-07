package worker

import "time"

const (
	StatusCreating = "creating"
	StatusRunning  = "running"
	StatusStopped  = "stopped"
	StatusError    = "error"
)

type SSH struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	User string `json:"user"`
}

type Box struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Owner     string    `json:"owner"`
	Status    string    `json:"status"`
	CPUs      int       `json:"cpus"`
	Memory    string    `json:"memory"`
	Disk      string    `json:"disk"`
	Net       string    `json:"net"`
	Error     string    `json:"error,omitempty"`
	SSH       SSH       `json:"ssh"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateRequest struct {
	Name   string `json:"name"`
	Owner  string `json:"owner"`
	CPUs   int    `json:"cpus"`
	Memory string `json:"memory"`
	Disk   string `json:"disk"`
	SSHKey string `json:"ssh_key,omitempty"`
	Net    string `json:"net,omitempty"`
}

type Snapshot struct {
	Tag       string    `json:"tag"`
	CreatedAt time.Time `json:"created_at,omitzero"`
}
