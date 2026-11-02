package suite

import "context"

type opusTasksInput struct {
	Workspace string `json:"workspace,omitempty" jsonschema:"description=Workspace id or name; defaults to the first workspace"`
	Project   string `json:"project,omitempty" jsonschema:"description=Project id, name or slug to list every task of; omit for the tasks assigned to the user"`
	Status    string `json:"status,omitempty" jsonschema:"description=Status slug such as todo or done; done tasks are hidden unless a status is set"`
	Limit     int    `json:"limit,omitempty" jsonschema:"description=Maximum tasks (default 20, max 50)"`
}

type sablierStatusInput struct{}

func opusTools(set *toolSet, o Opus) {
	readTool(set, "opus_tasks", "List tasks from Opus project management: the user's own assigned tasks, or a whole project.",
		func(ctx context.Context, in opusTasksInput) (string, error) {
			query := OpusQuery{Workspace: in.Workspace, Project: in.Project, Status: in.Status, Limit: in.Limit}
			return jsonResult(o.Tasks(ctx, set.userID, query))
		})
}

func sablierTools(set *toolSet, s Sablier) {
	readTool(set, "sablier_status", "Show the user's currently running timer in Sablier.",
		func(ctx context.Context, _ sablierStatusInput) (string, error) {
			return rawResult(s.Running(ctx, set.userID))
		})
}
