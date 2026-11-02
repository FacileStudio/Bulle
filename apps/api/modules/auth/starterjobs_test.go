package auth

import (
	"testing"

	"github.com/robfig/cron/v3"
)

func TestStarterJobSchedulesParse(t *testing.T) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	for _, job := range starterJobs {
		if _, err := parser.Parse(job.schedule); err != nil {
			t.Fatalf("%s: %v", job.name, err)
		}
	}
}
