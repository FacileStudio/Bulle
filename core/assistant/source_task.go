package assistant

import (
	"context"
	"strconv"

	"github.com/FacileStudio/Bulle/core/settings"
)

// TaskSource reads open tasks through the opus CLI.
type TaskSource struct {
	Suite    Suite
	Settings settings.Assistant
}

// Fetch returns the open tasks of every picked workspace. Done tasks are left
// out.
func (t *TaskSource) Fetch(ctx context.Context) ([]Item, error) {
	spaces, err := t.Suite.Workspaces(ctx)
	if err != nil {
		return nil, err
	}
	var items []Item
	for _, ws := range spaces {
		n, err := strconv.Atoi(ws.ID)
		if err != nil || !pickedID(t.Settings.Sources.Workspaces, n) {
			continue
		}
		got, err := t.tasks(ctx, ws)
		if err != nil {
			return nil, err
		}
		items = append(items, got...)
	}
	return items, nil
}

func (t *TaskSource) tasks(ctx context.Context, ws Workspace) ([]Item, error) {
	rows, err := t.Suite.Tasks(ctx, ws.ID)
	if err != nil {
		return nil, err
	}
	var items []Item
	for _, r := range rows {
		if r.Done || r.Status == "done" {
			continue
		}
		it := Item{Source: SourceTask, ID: r.ID, Title: r.Title, Account: ws.Name, Status: r.Status}
		if r.DueDate != nil {
			it.Due = *r.DueDate
		}
		items = append(items, it)
	}
	return items, nil
}
