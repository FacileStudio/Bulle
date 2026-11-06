package suite

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	opusTaskPath     = "/api/task/%d"
	opusMoveTaskPath = "/api/task/move/%d"
	opusCreatePath   = "/api/task/%d"
	opusProjectPath  = "/api/project/%d"
)

var opusWritableKeys = []string{
	"title", "description", "status", "priority", "projectId", "position", "startDate", "dueDate", "userId", "actorId",
}

var opusProjectKeys = []string{"id", "workspaceId", "slug", "icon", "name", "description", "isPublic"}

var opusTaskKeys = []string{"id", "projectId", "number", "title", "description", "status", "priority", "startDate", "dueDate", "actorName"}

type OpusProject struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type OpusProjectFields struct {
	Name *string `json:"name,omitempty" jsonschema:"description=Project name; required when creating"`
	Icon *string `json:"icon,omitempty"`
	Slug *string `json:"slug,omitempty"`
}

type OpusTaskFields struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty" jsonschema:"description=Markdown description"`
	Status      *string `json:"status,omitempty" jsonschema:"description=Column slug such as todo, in_progress or done"`
	Priority    *string `json:"priority,omitempty" jsonschema:"enum=none,enum=low,enum=medium,enum=high,enum=urgent"`
	StartDate   *string `json:"startDate,omitempty" jsonschema:"description=RFC 3339 or YYYY-MM-DD; empty clears it"`
	DueDate     *string `json:"dueDate,omitempty" jsonschema:"description=RFC 3339 or YYYY-MM-DD; empty clears it"`
}

func (o Opus) Projects(ctx context.Context, userID int64, workspaceRef string) ([]OpusProject, error) {
	workspace, err := o.workspace(ctx, userID, workspaceRef)
	if err != nil {
		return nil, err
	}
	all, err := o.projects(ctx, userID, workspace)
	if err != nil {
		return nil, err
	}
	out := make([]OpusProject, 0, len(all))
	for _, item := range all {
		out = append(out, OpusProject(item))
	}
	return out, nil
}

func (o Opus) Task(ctx context.Context, userID, taskID int64) ([]byte, error) {
	raw, err := o.gw.Get(ctx, userID, AppOpus, fmt.Sprintf(opusTaskPath, taskID), nil)
	if err != nil {
		return nil, err
	}
	return pickOne(raw, opusTaskKeys...)
}

func (o Opus) Create(ctx context.Context, userID int64, workspaceRef, projectRef string, fields OpusTaskFields) ([]byte, error) {
	if fields.Title == nil || *fields.Title == "" || projectRef == "" {
		return nil, errors.New("a task needs a title and a project")
	}
	workspace, err := o.workspace(ctx, userID, workspaceRef)
	if err != nil {
		return nil, err
	}
	project, err := o.project(ctx, userID, workspace, projectRef)
	if err != nil {
		return nil, err
	}
	body, err := opusPayload(fields, nil)
	if err != nil {
		return nil, err
	}
	raw, err := o.gw.Do(ctx, userID, Request{App: AppOpus, Method: http.MethodPost, Path: fmt.Sprintf(opusCreatePath, project.ID), Body: body})
	if err != nil {
		return nil, err
	}
	return pickOne(raw, opusTaskKeys...)
}

func (o Opus) Update(ctx context.Context, userID, taskID int64, projectRef string, fields OpusTaskFields) ([]byte, error) {
	path := fmt.Sprintf(opusTaskPath, taskID)
	if projectRef != "" {
		if err := o.move(ctx, userID, taskID, projectRef); err != nil {
			return nil, err
		}
	}
	current, err := o.gw.Get(ctx, userID, AppOpus, path, nil)
	if err != nil {
		return nil, err
	}
	stored, err := pickMap(current, opusWritableKeys...)
	if err != nil {
		return nil, fmt.Errorf("opus: reading task %d: %w", taskID, err)
	}
	body, err := opusPayload(fields, stored)
	if err != nil {
		return nil, err
	}
	raw, err := o.gw.Do(ctx, userID, Request{App: AppOpus, Method: http.MethodPut, Path: path, Body: body})
	if err != nil {
		return nil, err
	}
	return pickOne(raw, opusTaskKeys...)
}

func (o Opus) Project(ctx context.Context, userID int64, workspaceRef string, projectID int64) ([]byte, error) {
	workspace, err := o.workspace(ctx, userID, workspaceRef)
	if err != nil {
		return nil, err
	}
	current, err := o.gw.Get(ctx, userID, AppOpus, opusProjectsPath, url.Values{"workspaceId": {strconv.FormatInt(workspace, 10)}})
	if err != nil {
		return nil, err
	}
	var all []opusNamed
	if err := decodeInto(current, &all); err != nil {
		return nil, fmt.Errorf("opus: reading projects: %w", err)
	}
	var found *opusNamed
	for i := range all {
		if all[i].ID == projectID {
			found = &all[i]
			break
		}
	}
	if found == nil {
		return nil, fmt.Errorf("opus: no project with id %d", projectID)
	}
	raw, err := o.gw.Get(ctx, userID, AppOpus, fmt.Sprintf(opusProjectPath, found.ID), nil)
	if err != nil {
		return nil, err
	}
	return pickOne(raw, opusProjectKeys...)
}

func (o Opus) CreateProject(ctx context.Context, userID int64, workspaceRef string, fields OpusProjectFields) ([]byte, error) {
	if fields.Name == nil || *fields.Name == "" {
		return nil, errors.New("a project needs a name")
	}
	workspace, err := o.workspace(ctx, userID, workspaceRef)
	if err != nil {
		return nil, err
	}
	body := map[string]any{"name": *fields.Name, "workspaceId": workspace}
	if fields.Icon != nil && *fields.Icon != "" {
		body["icon"] = *fields.Icon
	}
	if fields.Slug != nil && *fields.Slug != "" {
		body["slug"] = *fields.Slug
	}
	raw, err := o.gw.Do(ctx, userID, Request{App: AppOpus, Method: http.MethodPost, Path: opusProjectsPath, Body: body})
	if err != nil {
		return nil, err
	}
	return pickOne(raw, opusProjectKeys...)
}

func (o Opus) UpdateProject(ctx context.Context, userID, projectID int64, fields OpusProjectFields) ([]byte, error) {
	path := fmt.Sprintf(opusProjectPath, projectID)
	body := map[string]any{}
	if fields.Name != nil {
		body["name"] = *fields.Name
	}
	if fields.Icon != nil {
		body["icon"] = *fields.Icon
	}
	if fields.Slug != nil {
		body["slug"] = *fields.Slug
	}
	if len(body) == 0 {
		return nil, errors.New("pass at least one of name, icon or slug")
	}
	raw, err := o.gw.Do(ctx, userID, Request{App: AppOpus, Method: http.MethodPut, Path: path, Body: body})
	if err != nil {
		return nil, err
	}
	return pickOne(raw, opusProjectKeys...)
}

func (o Opus) DeleteProject(ctx context.Context, userID, projectID int64) error {
	_, err := o.gw.Do(ctx, userID, Request{App: AppOpus, Method: http.MethodDelete, Path: fmt.Sprintf(opusProjectPath, projectID), Body: nil})
	return err
}

func (o Opus) Delete(ctx context.Context, userID, taskID int64) error {
	_, err := o.gw.Do(ctx, userID, Request{App: AppOpus, Method: http.MethodDelete, Path: fmt.Sprintf(opusTaskPath, taskID), Body: nil})
	return err
}

func (o Opus) move(ctx context.Context, userID, taskID int64, projectRef string) error {
	current, err := o.gw.Get(ctx, userID, AppOpus, fmt.Sprintf(opusTaskPath, taskID), nil)
	if err != nil {
		return err
	}
	var task struct {
		WorkspaceID int64 `json:"workspaceId"`
	}
	if err := decodeInto(current, &task); err != nil {
		return fmt.Errorf("opus: reading task %d: %w", taskID, err)
	}
	project, err := o.project(ctx, userID, task.WorkspaceID, projectRef)
	if err != nil {
		return err
	}
	_, err = o.gw.Do(ctx, userID, Request{App: AppOpus, Method: http.MethodPut, Path: fmt.Sprintf(opusMoveTaskPath, taskID), Body: map[string]any{"destinationProjectId": project.ID}})
	return err
}

func opusPayload(fields OpusTaskFields, stored map[string]any) (map[string]any, error) {
	body := map[string]any{}
	for key, value := range stored {
		body[key] = value
	}
	patch, err := toMap(fields)
	if err != nil {
		return nil, err
	}
	for key, value := range patch {
		body[key] = value
	}
	for key, value := range map[string]*string{"startDate": fields.StartDate, "dueDate": fields.DueDate} {
		if value == nil {
			continue
		}
		if *value == "" {
			delete(body, key)
			continue
		}
		parsed, err := parseDay(*value)
		if err != nil {
			return nil, err
		}
		body[key] = parsed
	}
	return body, nil
}

func parseDay(value string) (time.Time, error) {
	now := time.Now().UTC()
	return parseWhen(value, time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC), time.UTC)
}
