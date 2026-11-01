package computers

import "time"

const (
	StatusAbsent   = "absent"
	StatusCreating = "creating"
	StatusRunning  = "running"
	StatusStopped  = "stopped"
	StatusError    = "error"
)

type Computer struct {
	ID           int64     `gorm:"column:id;primaryKey" json:"-"`
	UserID       int64     `gorm:"column:user_id;uniqueIndex" json:"-"`
	BoxID        string    `gorm:"column:box_id" json:"-"`
	Status       string    `gorm:"column:status" json:"status"`
	SSHHost      string    `gorm:"column:ssh_host" json:"-"`
	SSHPort      int       `gorm:"column:ssh_port" json:"-"`
	CPUs         int       `gorm:"column:cpus" json:"cpus"`
	MemoryMB     int       `gorm:"column:memory_mb" json:"memory_mb"`
	DiskGB       int       `gorm:"column:disk_gb" json:"disk_gb"`
	LastActiveAt time.Time `gorm:"column:last_active_at" json:"last_active_at"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"-"`
	SSHUser      string    `gorm:"-" json:"-"`
	Error        string    `gorm:"-" json:"error,omitempty"`
}

func (Computer) TableName() string { return "computers" }
