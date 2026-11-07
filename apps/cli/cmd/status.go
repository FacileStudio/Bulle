package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

type statusLevel int

const (
	statusOK statusLevel = iota
	statusWarn
	statusFail
)

type statusLine struct {
	level  statusLevel
	name   string
	detail string
}

func (l statusLine) label() string {
	switch l.level {
	case statusFail:
		return "FAIL"
	case statusWarn:
		return "warn"
	}
	return "ok"
}

func newStatusCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check config, chat daemon, cron entries, jobs and the mail triage in one go",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			lines := collectStatus(c, version)
			failed, err := printStatus(c.OutOrStdout(), lines)
			if err != nil {
				return err
			}
			if failed {
				return fmt.Errorf("status: at least one check failed")
			}
			return nil
		},
	}
}

func printStatus(w io.Writer, lines []statusLine) (bool, error) {
	failed := false
	for _, l := range lines {
		if _, err := fmt.Fprintf(w, "%-4s  %-14s %s\n", l.label(), l.name, l.detail); err != nil {
			return failed, err
		}
		failed = failed || l.level == statusFail
	}
	return failed, nil
}

func collectStatus(c *cobra.Command, version string) []statusLine {
	lines := []statusLine{{statusOK, "version", version}}
	lines = append(lines, checkConfig()...)
	lines = append(lines, checkServer(c.Context())...)
	return lines
}

func homeFile(name string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home + "/" + name
}
