package suite

import (
	"context"
)

type opusTasksInput struct {
	Workspace string `json:"workspace,omitempty" jsonschema:"description=Workspace id or name; defaults to the first workspace"`
	Project   string `json:"project,omitempty" jsonschema:"description=Project id, name or slug to list every task of; omit for the tasks assigned to the user"`
	Status    string `json:"status,omitempty" jsonschema:"description=Status slug such as todo or done; done tasks are hidden unless a status is set"`
	Limit     int    `json:"limit,omitempty" jsonschema:"description=Maximum tasks (default 20, max 50)"`
}

type opusProjectsInput struct {
	Workspace string `json:"workspace,omitempty" jsonschema:"description=Workspace id or name; defaults to the first workspace"`
}

type opusProjectInput struct {
	ID int64 `json:"id" jsonschema:"required,description=Project id"`
}

type opusCreateProjectInput struct {
	Workspace string `json:"workspace,omitempty" jsonschema:"description=Workspace id or name; defaults to the first workspace"`
	OpusProjectFields
}

type opusUpdateProjectInput struct {
	ID int64 `json:"id" jsonschema:"required,description=Project id"`
	OpusProjectFields
}

type opusTaskInput struct {
	ID int64 `json:"id" jsonschema:"required,description=Global task id (the id field, not the per-project number)"`
}

type opusCreateInput struct {
	Workspace string `json:"workspace,omitempty" jsonschema:"description=Workspace id or name; defaults to the first workspace"`
	Project   string `json:"project" jsonschema:"required,description=Project id, name or slug"`
	OpusTaskFields
}

type opusUpdateInput struct {
	ID      int64  `json:"id" jsonschema:"required,description=Global task id (the id field, not the per-project number)"`
	Project string `json:"project,omitempty" jsonschema:"description=Move the task to this project id, name or slug"`
	OpusTaskFields
}

type sablierStatusInput struct{}

type sablierProjectsInput struct {
	Project int64 `json:"project,omitempty" jsonschema:"description=Project id to list the tasks of; omit to list projects"`
}

type sablierStartInput struct {
	Project int64 `json:"project" jsonschema:"required,description=Sablier project id"`
	Task    int64 `json:"task" jsonschema:"required,description=Sablier task id inside that project"`
}

func opusTools(set *toolSet, o Opus) {
	readTool(set, "opus_tasks", "List tasks from Opus project management: the user's own assigned tasks, or a whole project.",
		func(ctx context.Context, in opusTasksInput) (string, error) {
			query := OpusQuery{Workspace: set.defaults.workspace(in.Workspace), Project: in.Project, Status: in.Status, Limit: in.Limit}
			return jsonResult(o.Tasks(ctx, set.userID, query))
		})
	readTool(set, "opus_projects", "List the projects of an Opus workspace.",
		func(ctx context.Context, in opusProjectsInput) (string, error) {
			return jsonResult(o.Projects(ctx, set.userID, set.defaults.workspace(in.Workspace)))
		})
	readTool(set, "opus_task", "Show one Opus task with its description.",
		func(ctx context.Context, in opusTaskInput) (string, error) {
			return rawResult(o.Task(ctx, set.userID, in.ID))
		})
	readTool(set, "opus_project", "Show one Opus project with its description.",
		func(ctx context.Context, in opusProjectInput) (string, error) {
			return rawResult(o.Project(ctx, set.userID, set.defaults.workspace(""), in.ID))
		})
	writeTool(set, "opus_create_project", "Create a project in an Opus workspace.",
		func(ctx context.Context, in opusCreateProjectInput) (string, error) {
			return rawResult(o.CreateProject(ctx, set.userID, in.Workspace, in.OpusProjectFields))
		})
	writeTool(set, "opus_update_project", "Change an Opus project: name, icon or slug. Only the fields you pass change.",
		func(ctx context.Context, in opusUpdateProjectInput) (string, error) {
			return rawResult(o.UpdateProject(ctx, set.userID, in.ID, in.OpusProjectFields))
		})
	writeTool(set, "opus_delete_project", "Delete an Opus project.",
		func(ctx context.Context, in opusProjectInput) (string, error) {
			return done("deleted", o.DeleteProject(ctx, set.userID, in.ID))
		})
	writeTool(set, "opus_create_task", "Create a task in an Opus project.",
		func(ctx context.Context, in opusCreateInput) (string, error) {
			return rawResult(o.Create(ctx, set.userID, set.defaults.workspace(in.Workspace), in.Project, in.OpusTaskFields))
		})
	writeTool(set, "opus_update_task", "Change an Opus task: title, description, status (set done to complete it), priority, dates or project. Only the fields you pass change.",
		func(ctx context.Context, in opusUpdateInput) (string, error) {
			return rawResult(o.Update(ctx, set.userID, in.ID, in.Project, in.OpusTaskFields))
		})
	writeTool(set, "opus_delete_task", "Delete an Opus task.",
		func(ctx context.Context, in opusTaskInput) (string, error) {
			return done("deleted", o.Delete(ctx, set.userID, in.ID))
		})
}

func sablierTools(set *toolSet, s Sablier) {
	readTool(set, "sablier_status", "Show the user's currently running timer in Sablier.",
		func(ctx context.Context, _ sablierStatusInput) (string, error) {
			return rawResult(s.Running(ctx, set.userID))
		})
	readTool(set, "sablier_projects", "List the user's Sablier projects, or the tasks of one project.",
		func(ctx context.Context, in sablierProjectsInput) (string, error) {
			return rawResult(s.Projects(ctx, set.userID, in.Project))
		})
	writeTool(set, "sablier_start", "Start a Sablier timer on a task. Sablier refuses while another timer runs, so stop that one first.",
		func(ctx context.Context, in sablierStartInput) (string, error) {
			return rawResult(s.Start(ctx, set.userID, in.Project, in.Task))
		})
	writeTool(set, "sablier_stop", "Stop the user's running Sablier timer.",
		func(ctx context.Context, _ sablierStatusInput) (string, error) {
			return rawResult(s.Stop(ctx, set.userID))
		})
}
