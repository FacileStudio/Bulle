package auth

import (
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/suite"

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

func TestGloutonStarterOnlyWhenGloutonIsEnabled(t *testing.T) {
	names := func(jobs []starterJob) string {
		var out []string
		for _, job := range jobs {
			out = append(out, job.name)
		}
		return strings.Join(out, ",")
	}
	without, with := names(startersFor(nil)), names(startersFor([]string{suite.AppCourrier, suite.AppGlouton}))
	if strings.Contains(without, "Glouton") || !strings.Contains(with, "Glouton daily review") || !strings.Contains(without, "Morning brief") {
		t.Fatalf("without Glouton %q, with Glouton %q", without, with)
	}
}
