package suite

import "context"

type sablierStatusInput struct{}

type sablierProjectsInput struct {
	Project int64 `json:"project,omitempty" jsonschema:"description=Project id to list the tasks of; omit to list projects"`
}

type sablierStartInput struct {
	Project int64 `json:"project" jsonschema:"required,description=Sablier project id"`
	Task    int64 `json:"task" jsonschema:"required,description=Sablier task id inside that project"`
}

func sablierTools(set *toolSet, s Sablier) {
	readTool(set, "sablier_status", "Show the user's currently running Sablier timer, if any. "+
		"Use it before starting a timer, since only one runs at a time.",
		func(ctx context.Context, _ sablierStatusInput) (string, error) {
			return rawResult(s.Running(ctx, set.userID))
		})
	readTool(set, "sablier_projects", "List the user's Sablier projects, or the tasks of one project when project is set. "+
		"Use it to find the project and task ids sablier_start takes.",
		func(ctx context.Context, in sablierProjectsInput) (string, error) {
			return rawResult(s.Projects(ctx, set.userID, in.Project))
		})
	writeTool(set, "sablier_start", "Start a Sablier timer on a task. Sablier refuses while another timer runs, so stop that one first with sablier_stop. "+
		"project and task are ids from sablier_projects.",
		func(ctx context.Context, in sablierStartInput) (string, error) {
			return rawResult(s.Start(ctx, set.userID, in.Project, in.Task))
		})
	writeTool(set, "sablier_stop", "Stop the user's running Sablier timer and save the time entry. It takes no arguments.",
		func(ctx context.Context, _ sablierStatusInput) (string, error) {
			return rawResult(s.Stop(ctx, set.userID))
		})
}
