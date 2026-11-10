package auth

import (
	"slices"
	"strconv"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/tronc/errors"
	"gorm.io/gorm"
)

type starterJob struct {
	name     string
	schedule string
	prompt   string
	app      string
}

const briefRules = `Use your mail, calendar and task tools: courrier_search_mail with unread_only for unread mail, agenda_list_events for the calendar, ` +
	`opus_list_tasks for open tasks. If a tool is not connected or fails, skip that part in one short line and keep going. ` +
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
	{
		name:     "Glouton daily review",
		schedule: "CRON_TZ=Europe/Paris 30 9 * * 1-5",
		app:      suite.AppGlouton,
		prompt: "Run my daily prospecting review in Glouton. " +
			"First read glouton_list_suggestions so you do not repeat an open suggestion. " +
			"Then use glouton_search_contacts with temperature HOT, then WARM, to find prospects nobody has contacted yet, " +
			"read each with glouton_get_company or glouton_get_person, and save a short, specific first-contact draft with glouton_save_draft. " +
			"Next find follow-ups due: use glouton_list_outreach on contacted prospects and pick outreach SENT more than 5 days ago with no reply; " +
			"save one follow-up draft for each with glouton_save_draft. " +
			"When the evidence says a temperature or score is wrong, call glouton_suggest with kind RESCORE; " +
			"when a contact needs a step other than an email, call glouton_suggest with kind NEXT_ACTION. " +
			"Give every draft and suggestion a one-line rationale. Never invent a fact about a company or a person. " +
			"Drafts and suggestions go straight to Glouton's agent inbox without an approval card here, and nothing is sent: " +
			"I review them and send or accept them myself there. Contact edits still ask for my approval, so keep them rare. " +
			"End with a short summary: drafts written, follow-ups, suggestions, and the inbox link. " +
			"If Glouton is not connected or fails, say so in one line and stop.",
	},
}

func startersFor(apps []string) []starterJob {
	var jobs []starterJob
	for _, job := range starterJobs {
		if job.app == "" || slices.Contains(apps, job.app) {
			jobs = append(jobs, job)
		}
	}
	return jobs
}

func seedStarterJobs(tx *gorm.DB, userID int64, apps []string) error {
	uid := strconv.FormatInt(userID, 10)
	spaceID := store.NewID()
	err := tx.Exec("INSERT INTO spaces (id, user_id, name) VALUES (?, ?, ?)", spaceID, userID, store.DefaultSpaceName).Error
	if err != nil {
		return errors.Internal("failed to create the default space", err)
	}
	for _, job := range startersFor(apps) {
		err := tx.Exec(
			"INSERT INTO jobs (id, user_id, space_id, name, schedule, prompt, delivery, enabled) VALUES (?, ?, ?, ?, ?, ?, ?, true)",
			store.NewID(), uid, spaceID, job.name, job.schedule, job.prompt, "push",
		).Error
		if err != nil {
			return errors.Internal("failed to create the starter jobs", err)
		}
	}
	return nil
}
