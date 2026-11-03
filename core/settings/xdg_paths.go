package settings

import (
	"fmt"
	"os"
	"path/filepath"
)

// LogDir is where cron jobs and delivery fallbacks write their logs: logs/
// inside StateDir, or "" when no home directory resolves.
func LogDir() string {
	dir, err := StateDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "logs")
}

func configDir() (string, error) { return xdgDir("XDG_CONFIG_HOME", ".config") }

func dataDir() (string, error) {
	return xdgDir("XDG_DATA_HOME", filepath.Join(".local", "share"))
}

func stateDir() (string, error) {
	return xdgDir("XDG_STATE_HOME", filepath.Join(".local", "state"))
}

func xdgDir(env, fallback string) (string, error) {
	if dir := absoluteEnv(env); dir != "" {
		return filepath.Join(dir, xdgApp), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving %s: %w", env, err)
	}
	return filepath.Join(home, fallback, xdgApp), nil
}

func absoluteEnv(name string) string {
	dir := os.Getenv(name)
	if dir == "" || !filepath.IsAbs(dir) {
		return ""
	}
	return dir
}
