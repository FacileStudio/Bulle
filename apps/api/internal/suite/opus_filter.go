package suite

import (
	"context"
	"slices"
)

// statusFilter resolves one status against the columns of every project a
// task list spans, fetching each project's columns once.
type statusFilter struct {
	opus   Opus
	userID int64
	status string
	slugs  map[int64]string
	valid  []string
}

func (f *statusFilter) slug(ctx context.Context, projectID int64) (string, error) {
	if slug, ok := f.slugs[projectID]; ok {
		return slug, nil
	}
	columns, err := f.opus.columns(ctx, f.userID, projectID)
	if err != nil {
		return "", err
	}
	slug, _ := matchStatus(columns, f.status)
	f.slugs[projectID] = slug
	for _, valid := range opusStatuses(columns) {
		if !slices.Contains(f.valid, valid) {
			f.valid = append(f.valid, valid)
		}
	}
	return slug, nil
}

// keepStatus keeps the tasks whose column is the one status names in their
// own project. A status no project knows is an error listing the valid ones,
// never an empty list.
func (o Opus) keepStatus(ctx context.Context, userID int64, tasks []OpusTask, status string) ([]OpusTask, error) {
	filter := &statusFilter{opus: o, userID: userID, status: status, slugs: map[int64]string{}}
	kept := []OpusTask{}
	matched := false
	for _, task := range tasks {
		slug, err := filter.slug(ctx, task.ProjectID)
		if err != nil {
			return nil, err
		}
		matched = matched || slug != ""
		if slug != "" && task.Status == slug {
			kept = append(kept, task)
		}
	}
	if len(tasks) > 0 && !matched {
		return nil, statusError(status, filter.valid)
	}
	return kept, nil
}
