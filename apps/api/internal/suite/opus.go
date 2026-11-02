package suite

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	opusWorkspacesPath = "/api/workspace"
	opusProjectsPath   = "/api/project"
	opusMyTasksPath    = "/api/task/me"
	opusTasksPath      = "/api/task/tasks/%d"
)

type opusNamed struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type OpusTask struct {
	ID        int64      `json:"id"`
	ProjectID int64      `json:"projectId"`
	Project   string     `json:"project,omitempty"`
	Title     string     `json:"title"`
	Status    string     `json:"status"`
	Priority  string     `json:"priority,omitempty"`
	DueDate   *time.Time `json:"dueDate,omitempty"`
	Assignee  string     `json:"assignee,omitempty"`
}

type OpusQuery struct {
	Workspace string
	Project   string
	Status    string
	Limit     int
}

type Opus struct{ gw *Gateway }

func NewOpus(gw *Gateway) Opus { return Opus{gw: gw} }

func (o Opus) Tasks(ctx context.Context, userID int64, q OpusQuery) ([]OpusTask, error) {
	workspace, err := o.workspace(ctx, userID, q.Workspace)
	if err != nil {
		return nil, err
	}
	var tasks []OpusTask
	if q.Project == "" {
		tasks, err = o.myTasks(ctx, userID, workspace, q.Status)
	} else {
		tasks, err = o.projectTasks(ctx, userID, workspace, q)
	}
	if err != nil {
		return nil, err
	}
	return tasks[:min(len(tasks), clampLimit(q.Limit, 20))], nil
}

func (o Opus) workspace(ctx context.Context, userID int64, ref string) (int64, error) {
	raw, err := o.gw.Get(ctx, userID, AppOpus, opusWorkspacesPath, nil)
	if err != nil {
		return 0, err
	}
	var all []opusNamed
	if err := decodeInto(raw, &all); err != nil {
		return 0, fmt.Errorf("opus: reading workspaces: %w", err)
	}
	if ref == "" && len(all) > 0 {
		return all[0].ID, nil
	}
	found, err := filterNamed(all, ref)
	if err != nil {
		return 0, err
	}
	return found[0].ID, nil
}

func (o Opus) myTasks(ctx context.Context, userID, workspace int64, status string) ([]OpusTask, error) {
	query := url.Values{"workspaceId": {strconv.FormatInt(workspace, 10)}}
	if status != "" {
		query.Set("includeDone", "true")
	}
	raw, err := o.gw.Get(ctx, userID, AppOpus, opusMyTasksPath, query)
	if err != nil {
		return nil, err
	}
	var mine []struct {
		OpusTask
		ActorName   string `json:"actorName"`
		ProjectName string `json:"projectName"`
	}
	if err := decodeInto(raw, &mine); err != nil {
		return nil, fmt.Errorf("opus: reading your tasks: %w", err)
	}
	tasks := []OpusTask{}
	for _, item := range mine {
		task := item.OpusTask
		task.Project, task.Assignee = item.ProjectName, item.ActorName
		if status == "" || strings.EqualFold(task.Status, status) {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (o Opus) projectTasks(ctx context.Context, userID, workspace int64, q OpusQuery) ([]OpusTask, error) {
	project, err := o.project(ctx, userID, workspace, q.Project)
	if err != nil {
		return nil, err
	}
	return o.board(ctx, userID, project, q.Status)
}

func (o Opus) projects(ctx context.Context, userID, workspace int64) ([]opusNamed, error) {
	raw, err := o.gw.Get(ctx, userID, AppOpus, opusProjectsPath, url.Values{"workspaceId": {strconv.FormatInt(workspace, 10)}})
	if err != nil {
		return nil, err
	}
	var all []opusNamed
	if err := decodeInto(raw, &all); err != nil {
		return nil, fmt.Errorf("opus: reading projects: %w", err)
	}
	return all, nil
}

func (o Opus) project(ctx context.Context, userID, workspace int64, ref string) (opusNamed, error) {
	all, err := o.projects(ctx, userID, workspace)
	if err != nil {
		return opusNamed{}, err
	}
	found, err := filterNamed(all, ref)
	if err != nil {
		return opusNamed{}, err
	}
	return found[0], nil
}

func (o Opus) board(ctx context.Context, userID int64, project opusNamed, status string) ([]OpusTask, error) {
	query := url.Values{}
	if status != "" {
		query.Set("status", status)
	}
	raw, err := o.gw.Get(ctx, userID, AppOpus, fmt.Sprintf(opusTasksPath, project.ID), query)
	if err != nil {
		return nil, err
	}
	var board struct {
		Columns []struct {
			IsFinal bool `json:"isFinal"`
			Tasks   []struct {
				OpusTask
				ActorName string `json:"actorName"`
			} `json:"tasks"`
		} `json:"columns"`
	}
	if err := json.Unmarshal(unwrapData(raw), &board); err != nil {
		return nil, fmt.Errorf("opus: reading tasks: %w", err)
	}
	var tasks []OpusTask
	for _, column := range board.Columns {
		if column.IsFinal && status == "" {
			continue
		}
		for _, item := range column.Tasks {
			task := item.OpusTask
			task.Project, task.Assignee = project.Name, item.ActorName
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func filterNamed(items []opusNamed, ref string) ([]opusNamed, error) {
	ref = strings.TrimSpace(ref)
	var out []opusNamed
	for _, item := range items {
		if strconv.FormatInt(item.ID, 10) == ref || strings.EqualFold(item.Name, ref) || strings.EqualFold(item.Slug, ref) {
			out = append(out, item)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("opus: nothing is named %q", ref)
	}
	return out, nil
}
