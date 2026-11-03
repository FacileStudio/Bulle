package cmd

import (
	"os"

	"github.com/FacileStudio/Bulle/core/settings"
)

func checkConfig() []statusLine {
	lines := []statusLine{}
	if _, err := settings.Settings("", settings.Config{}); err != nil {
		lines = append(lines, statusLine{statusFail, "config", err.Error()})
	} else {
		lines = append(lines, statusLine{statusOK, "config", settings.ConfigPath()})
	}
	return append(lines, legacyLines()...)
}

func legacyLines() []statusLine {
	var lines []statusLine
	for _, legacy := range []string{".bulle", ".bulle.yml"} {
		path := homeFile(legacy)
		if path == "" {
			continue
		}
		if entries, err := os.ReadDir(path); err == nil && len(entries) > 0 || err != nil && !os.IsNotExist(err) {
			lines = append(lines, statusLine{statusWarn, "legacy", path + " still exists, bulle did not migrate it"})
		}
	}
	return lines
}
