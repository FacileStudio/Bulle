package computers

import (
	"fmt"
	"os"
	"strings"
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
}

func NewConfig(boite env.Boite) (Config, error) {
	keyPath, err := resolveKey(boite.BoxKey)
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
		KeyPath:      keyPath,
	}, nil
}

func resolveKey(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || !strings.Contains(value, "PRIVATE KEY") {
		return value, nil
	}
	file, err := os.CreateTemp("", "bulle-box-key-*")
	if err != nil {
		return "", fmt.Errorf("failed to stage the box key: %w", err)
	}
	defer file.Close()
	if err := file.Chmod(0o600); err != nil {
		return "", fmt.Errorf("failed to protect the box key: %w", err)
	}
	if _, err := file.WriteString(value + "\n"); err != nil {
		return "", fmt.Errorf("failed to stage the box key: %w", err)
	}
	return file.Name(), nil
}
