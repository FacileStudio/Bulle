package runner

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

const feedbackItems = 30

const ideasRules = `Rules:
- Propose 3 to 7 ideas with idea_propose, at most 7. When nothing is worth suggesting, propose nothing.
- Each idea is something you can do with your tools in a chat with the user: reply to a mail, plan an event, move a task, prepare research, follow up.
- Give a concrete "why now" grounded in something you read. Prefer what is due, what waits on the user, and what moves a goal.
- Never reuse a fingerprint listed above. Fingerprints are stable keys like courrier:5/email:812/reply or opus:task:44/due.
- Drafts are text only. Do not send, save, change or delete anything.
- No idea that only says "check your inbox" or "review your calendar".
- Write the prompt as the user's own request to you, complete enough to act on without this context.
- Write title, why, prompt and draft in the language the user writes to you in; follow the language of the steering text and goals when they have one.
- Read first, then propose, then end your turn with one short line. Do not write a report.`

func (r *Runner) ideasPrompt(ctx context.Context, run *ideaRun) (string, error) {
	st := r.cfg.Store
	userID, spaceID := run.batch.UserID, run.space.ID
	settings, err := st.GetIdeaSettings(ctx, userID, spaceID)
	if errors.Is(err, store.ErrNotFound) {
		settings, err = store.DefaultIdeaSettings(userID, spaceID), nil
	}
	if err != nil {
		return "", err
	}
	goals, err := st.ListGoals(ctx, userID, spaceID)
	if err != nil {
		return "", err
	}
	feedback, err := st.IdeaFeedback(ctx, userID, spaceID, feedbackItems)
	if err != nil {
		return "", err
	}
	taken, err := st.SuppressedFingerprints(ctx, userID, spaceID, r.now())
	if err != nil {
		return "", err
	}
	sections := []string{
		localNow(r.now(), settings.Timezone),
		"Space: " + run.space.Name + ".",
		"Open goals:\n" + goalLines(goals),
		"What the user wants you to think about:\n" + orNone(settings.Steering),
		"Feedback on recent ideas:\n" + feedbackLines(feedback),
		"Fingerprints already suggested or dismissed:\n" + orNone(strings.Join(taken, "\n")),
		ideasRules,
	}
	return strings.Join(sections, "\n\n"), nil
}

func localNow(now time.Time, timezone string) string {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc, timezone = time.UTC, "UTC"
	}
	return fmt.Sprintf("Today is %s, %s in %s.", now.In(loc).Format("Monday 2 January 2006"), now.In(loc).Format("15:04"), timezone)
}

func goalLines(goals []store.Goal) string {
	var lines []string
	for _, goal := range goals {
		if !goal.Done {
			lines = append(lines, fmt.Sprintf("- [%s] %s", goal.Category, goal.Title))
		}
	}
	return orNone(strings.Join(lines, "\n"))
}

func feedbackLines(ideas []store.Idea) string {
	lines := make([]string, 0, len(ideas))
	for _, idea := range ideas {
		var notes []string
		switch idea.Feedback {
		case store.FeedbackUp:
			notes = append(notes, "thumbs up")
		case store.FeedbackDown:
			notes = append(notes, "thumbs down")
		}
		notes = append(notes, idea.Status)
		if idea.Reason != "" {
			notes = append(notes, "reason: "+idea.Reason)
		}
		lines = append(lines, fmt.Sprintf("- %q (%s): %s", idea.Title, idea.Kind, strings.Join(notes, ", ")))
	}
	return orNone(strings.Join(lines, "\n"))
}

func orNone(text string) string {
	if strings.TrimSpace(text) == "" {
		return "none"
	}
	return text
}
