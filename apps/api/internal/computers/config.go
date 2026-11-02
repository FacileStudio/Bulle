package computers

import (
	"log/slog"
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

func NewConfig(boite env.Boite, logger *slog.Logger) (Config, error) {
	key, err := loadBoxKey(boite.BoxKey)
	if err != nil {
		// A key that cannot be read must not take the API down. Boxes are one
		// feature, and a pair generated here is published to each box on its next
		// run, so a misconfigured secret costs a warning and a new key, not a
		// restart loop on every deploy.
		logger.Warn("the configured box key is unusable, generating one for this boot", slog.Any("error", err))
		key, err = generateBoxKey()
		if err != nil {
			return Config{}, err
		}
	}
	idleAfter := time.Duration(boite.IdleMinutes) * time.Minute
	if idleAfter < 0 {
		idleAfter = 0
	}
	return Config{
		CPUs:         boite.BoxCPUs,
		MemoryMB:     boite.BoxMemoryMB,
		DiskGB:       boite.BoxDiskGB,
		IdleAfter:    idleAfter,
		BootTimeout:  3 * time.Minute,
		PollInterval: 2 * time.Second,
		KeyPath:      key.Path,
		PublicKey:    key.Public,
	}, nil
}
