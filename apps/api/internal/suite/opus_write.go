package suite

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	opusTaskPath     = "/api/task/%d"
	opusMoveTaskPath = "/api/task/move/%d"
	opusCreatePath   = "/api/task/%d"
)

var opusWritableKeys = []string{
	"title", "description", "status", "priority", "projectId", "position", "startDate", "dueDate", "userId", "actorId",
}

var opusTaskKeys = []string{"id", "facile_id", "projectId", "number", "title", "description", "status", "priority", "startDate", "dueDate", "actorName"}

// OpusTaskFields are the task fields the agent may set. A nil field is left
// untouched.
type OpusTaskFields struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty" jsonschema:"description=Markdown description"`
	Status      *string `json:"status,omitempty" jsonschema:"description=Column slug of the task's project such as to-do, in-progress, in-review or done; opus_get_project lists a project's columns"`
	Priority    *string `json:"priority,omitempty" jsonschema:"enum=none,enum=low,enum=medium,enum=high,enum=urgent"`
	StartDate   *string `json:"start_date,omitempty" jsonschema:"description=Start date, RFC 3339 or YYYY-MM-DD; an empty string clears it"`
	DueDate     *string `json:"due_date,omitempty" jsonschema:"description=Due date, RFC 3339 or YYYY-MM-DD; an empty string clears it"`
}

// Task shows one task by its global id.
func (o Opus) Task(ctx context.Context, userID, taskID int64) ([]byte, error) {
	raw, err := o.gw.Get(ctx, userID, AppOpus, fmt.Sprintf(opusTaskPath, taskID), nil)
	if err != nil {
		return nil, err
	}
	return pickOne(raw, opusTaskKeys...)
}

// Create creates a task in a project, its status resolved against the
// project's columns.
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
	if fields, err = o.withStatus(ctx, userID, project.ID, fields); err != nil {
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

// Update changes the fields passed, after moving the task when a project is
// named, its status resolved against the columns of the project it ends in.
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
	projectID, _ := stored["projectId"].(float64)
	if fields, err = o.withStatus(ctx, userID, int64(projectID), fields); err != nil {
		return nil, err
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

func (o Opus) withStatus(ctx context.Context, userID, projectID int64, fields OpusTaskFields) (OpusTaskFields, error) {
	if fields.Status == nil || *fields.Status == "" {
		return fields, nil
	}
	slug, err := o.resolveStatus(ctx, userID, projectID, *fields.Status)
	if err != nil {
		return fields, err
	}
	fields.Status = &slug
	return fields, nil
}

// Delete deletes a task by its global id.
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
	undated := fields
	undated.StartDate, undated.DueDate = nil, nil
	patch, err := toMap(undated)
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
