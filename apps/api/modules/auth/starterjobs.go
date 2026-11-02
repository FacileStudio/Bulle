package auth

import (
	"strconv"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/tronc/errors"
	"gorm.io/gorm"
)

type starterJob struct {
	name     string
	schedule string
	prompt   string
}

const briefRules = `Use your mail, calendar and task tools: courrier_inbox for unread mail, agenda_events for the calendar, ` +
	`opus_tasks for open tasks. If a tool is not connected or fails, skip that part in one short line and keep going. ` +
	`Never invent an email, event or task. Write in the user's language, in your own voice, short enough to read ` +
	`in under a minute. Lead with what needs action. Group by importance, not by source. ` +
	`If nothing needs attention, say so in one line.`

var starterJobs = []starterJob{
	{
		name:     "Morning brief",
		schedule: "CRON_TZ=Europe/Paris 0 8 * * 1-5",
		prompt: "Write my morning brief for today. Cover: today's meetings with their times, " +
			"unread emails that need a reply or a decision, and tasks due today or overdue. " +
			"End with the one thing I should start with. " + briefRules,
	},
	{
		name:     "Afternoon check-in",
		schedule: "CRON_TZ=Europe/Paris 0 14 * * 1-5",
		prompt: "Write my afternoon check-in. Cover: the rest of today's meetings, " +
			"emails that arrived since this morning and need an answer, and tasks due today that are still open. " +
			"End with what to finish before the day ends. " + briefRules,
	},
	{
		name:     "Evening recap",
		schedule: "CRON_TZ=Europe/Paris 0 19 * * 1-5",
		prompt: "Write my evening recap. Cover: what is still waiting on me in my inbox, " +
			"tasks that slipped today, and tomorrow's first meetings so I can prepare. " +
			"End with one line on what tomorrow looks like. " + briefRules,
	},
}

func seedStarterJobs(tx *gorm.DB, userID int64) error {
	uid := strconv.FormatInt(userID, 10)
	for _, job := range starterJobs {
		err := tx.Exec(
			"INSERT INTO jobs (id, user_id, name, schedule, prompt, delivery, enabled) VALUES (?, ?, ?, ?, ?, ?, true)",
			store.NewID(), uid, job.name, job.schedule, job.prompt, "push",
		).Error
		if err != nil {
			return errors.Internal("failed to create the starter jobs", err)
		}
	}
	return nil
}
