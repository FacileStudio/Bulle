package computers

import (
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/env"
)

type Config struct {
	CPUs         int
	MemoryMB     int
	DiskGB       int
	IdleAfter    time.Duration
	BootTimeout  time.Duration
	PollInterval time.Duration
	KeyPath      string
	PublicKey    string
}

func NewConfig(boite env.Boite) (Config, error) {
	key, err := loadBoxKey(boite.BoxKey)
	if err != nil {
		return Config{}, err
	}
	return Config{
		CPUs:         boite.BoxCPUs,
		MemoryMB:     boite.BoxMemoryMB,
		DiskGB:       boite.BoxDiskGB,
		IdleAfter:    time.Duration(boite.IdleMinutes) * time.Minute,
		BootTimeout:  3 * time.Minute,
		PollInterval: 2 * time.Second,
		KeyPath:      key.Path,
		PublicKey:    key.Public,
	}, nil
}
