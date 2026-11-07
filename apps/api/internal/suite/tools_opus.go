package suite

import (
	"context"
)

type opusTasksInput struct {
	Workspace string `json:"workspace,omitempty" jsonschema:"description=Workspace id or name; defaults to the conversation's workspace, else the first one"`
	Project   string `json:"project,omitempty" jsonschema:"description=Project id, name or slug to list every task of; omit for the tasks assigned to the user"`
	Status    string `json:"status,omitempty" jsonschema:"description=Column slug such as to-do, in-progress, in-review or done; done tasks are hidden unless a status is set"`
	Limit     int    `json:"limit,omitempty" jsonschema:"description=Maximum tasks (default 20, max 50)"`
}

type opusProjectsInput struct {
	Workspace string `json:"workspace,omitempty" jsonschema:"description=Workspace id or name; defaults to the conversation's workspace, else the first one"`
}

type opusShowProjectInput struct {
	Workspace string `json:"workspace,omitempty" jsonschema:"description=Workspace id or name the project is in; defaults to the conversation's workspace"`
	ProjectID int64  `json:"project_id" jsonschema:"required,description=Project id as opus_list_projects answers it"`
}

type opusProjectInput struct {
	ProjectID int64 `json:"project_id" jsonschema:"required,description=Project id as opus_list_projects answers it"`
}

type opusCreateProjectInput struct {
	Workspace string `json:"workspace,omitempty" jsonschema:"description=Workspace id or name; defaults to the conversation's workspace, else the first one"`
	OpusProjectFields
}

type opusUpdateProjectInput struct {
	ProjectID int64 `json:"project_id" jsonschema:"required,description=Project id as opus_list_projects answers it"`
	OpusProjectFields
}

type opusTaskInput struct {
	TaskID int64 `json:"task_id" jsonschema:"required,description=Global task id: the id field opus_list_tasks answers, not the per-project number"`
}

type opusCreateInput struct {
	Workspace string `json:"workspace,omitempty" jsonschema:"description=Workspace id or name; defaults to the conversation's workspace, else the first one"`
	Project   string `json:"project" jsonschema:"required,description=Project id, name or slug"`
	OpusTaskFields
}

type opusUpdateInput struct {
	TaskID  int64  `json:"task_id" jsonschema:"required,description=Global task id: the id field opus_list_tasks answers, not the per-project number"`
	Project string `json:"project,omitempty" jsonschema:"description=Move the task to this project id, name or slug"`
	OpusTaskFields
}

func opusTools(set *toolSet, o Opus) {
	opusReadTools(set, o)
	writeTool(set, "opus_create_project", "Create a project in an Opus workspace. Use it only when the user asks for a new project, never to file a single task. "+
		"name is required; icon and slug are optional.",
		func(ctx context.Context, in opusCreateProjectInput) (string, error) {
			return rawResult(o.CreateProject(ctx, set.userID, set.defaults.workspace(in.Workspace), in.OpusProjectFields))
		})
	writeTool(set, "opus_update_project", "Rename an Opus project or change its icon or slug. Only the fields you pass change. "+
		"project_id comes from opus_list_projects.",
		func(ctx context.Context, in opusUpdateProjectInput) (string, error) {
			return rawResult(o.UpdateProject(ctx, set.userID, in.ProjectID, in.OpusProjectFields))
		})
	writeTool(set, "opus_delete_project", "Delete an Opus project with every task in it. Use it only when the user explicitly asks to delete the project. "+
		"project_id comes from opus_list_projects.",
		func(ctx context.Context, in opusProjectInput) (string, error) {
			return done("deleted", o.DeleteProject(ctx, set.userID, in.ProjectID))
		})
	opusTaskWriteTools(set, o)
}

func opusReadTools(set *toolSet, o Opus) {
	readTool(set, "opus_list_tasks", "List Opus tasks: the ones assigned to the user, or every task of one project when project is set. "+
		"Use it to answer what is on the user's plate or to find a task's id before changing it. Done tasks are left out unless status asks for them. "+
		"status takes a column slug of the project; opus_get_project lists them.",
		func(ctx context.Context, in opusTasksInput) (string, error) {
			query := OpusQuery{Workspace: set.defaults.workspace(in.Workspace), Project: in.Project, Status: in.Status, Limit: in.Limit}
			return jsonResult(o.Tasks(ctx, set.userID, query))
		})
	readTool(set, "opus_list_projects", "List the projects of an Opus workspace with their id, name and slug. "+
		"Use it to find the project a task belongs in, or the project_id other Opus tools take.",
		func(ctx context.Context, in opusProjectsInput) (string, error) {
			return jsonResult(o.Projects(ctx, set.userID, set.defaults.workspace(in.Workspace)))
		})
	readTool(set, "opus_get_task", "Show one Opus task with its description, status, priority and dates. "+
		"Use it before changing a task, to read its current state. task_id is the global id opus_list_tasks answers.",
		func(ctx context.Context, in opusTaskInput) (string, error) {
			return rawResult(o.Task(ctx, set.userID, in.TaskID))
		})
	readTool(set, "opus_get_project", "Show one Opus project with its description and its board columns. "+
		"Each column's slug is a status its tasks can take, and the column marked isFinal is done. "+
		"Pass workspace when the project is outside the conversation's workspace.",
		func(ctx context.Context, in opusShowProjectInput) (string, error) {
			return rawResult(o.Project(ctx, set.userID, set.defaults.workspace(in.Workspace), in.ProjectID))
		})
}

func opusTaskWriteTools(set *toolSet, o Opus) {
	writeTool(set, "opus_create_task", "Create a task in an Opus project. project takes the project's id, name or slug. "+
		"status is a column slug of that project such as to-do or in-progress, and dates take RFC 3339 or YYYY-MM-DD. "+
		"Search opus_list_tasks first when the task may already exist.",
		func(ctx context.Context, in opusCreateInput) (string, error) {
			return rawResult(o.Create(ctx, set.userID, set.defaults.workspace(in.Workspace), in.Project, in.OpusTaskFields))
		})
	writeTool(set, "opus_update_task", "Change an Opus task: title, description, status, priority, dates, or the project it is in. "+
		"Only the fields you pass change. To complete a task, set status to done (or the project's final column). "+
		"task_id is the global id opus_list_tasks answers.",
		func(ctx context.Context, in opusUpdateInput) (string, error) {
			return rawResult(o.Update(ctx, set.userID, in.TaskID, in.Project, in.OpusTaskFields))
		})
	writeTool(set, "opus_delete_task", "Delete an Opus task for good. Use it only when the user asks to delete it; to finish a task, "+
		"set its status to done with opus_update_task instead. task_id is the global id opus_list_tasks answers.",
		func(ctx context.Context, in opusTaskInput) (string, error) {
			return done("deleted", o.Delete(ctx, set.userID, in.TaskID))
		})
}
