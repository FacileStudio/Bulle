package suite

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const opusProjectPath = "/api/project/%d"

var opusProjectKeys = []string{"id", "workspaceId", "slug", "icon", "name", "description", "isPublic"}

// OpusProject is one project as the project list shows it.
type OpusProject struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// OpusProjectFields are the project fields the agent may set. A nil field is
// left untouched.
type OpusProjectFields struct {
	Name *string `json:"name,omitempty" jsonschema:"description=Project name; required when creating"`
	Icon *string `json:"icon,omitempty"`
	Slug *string `json:"slug,omitempty"`
}

// Projects lists the projects of a workspace.
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

// Project shows one project of a workspace with its board columns, whose
// slugs are the statuses its tasks take.
func (o Opus) Project(ctx context.Context, userID int64, workspaceRef string, projectID int64) ([]byte, error) {
	workspace, err := o.workspace(ctx, userID, workspaceRef)
	if err != nil {
		return nil, err
	}
	all, err := o.projects(ctx, userID, workspace)
	if err != nil {
		return nil, err
	}
	if !projectListed(all, projectID) {
		return nil, fmt.Errorf("opus: no project with id %d in this workspace", projectID)
	}
	raw, err := o.gw.Get(ctx, userID, AppOpus, fmt.Sprintf(opusProjectPath, projectID), nil)
	if err != nil {
		return nil, err
	}
	project, err := pickMap(raw, opusProjectKeys...)
	if err != nil {
		return nil, fmt.Errorf("opus: reading project %d: %w", projectID, err)
	}
	columns, err := o.columns(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}
	project["columns"] = columns
	return json.Marshal(project)
}

func projectListed(all []opusNamed, projectID int64) bool {
	for _, item := range all {
		if item.ID == projectID {
			return true
		}
	}
	return false
}

// CreateProject creates a project in a workspace.
func (o Opus) CreateProject(ctx context.Context, userID int64, workspaceRef string, fields OpusProjectFields) ([]byte, error) {
	if fields.Name == nil || *fields.Name == "" {
		return nil, errors.New("a project needs a name")
	}
	workspace, err := o.workspace(ctx, userID, workspaceRef)
	if err != nil {
		return nil, err
	}
	body := projectBody(fields, true)
	body["workspaceId"] = workspace
	raw, err := o.gw.Do(ctx, userID, Request{App: AppOpus, Method: http.MethodPost, Path: opusProjectsPath, Body: body})
	if err != nil {
		return nil, err
	}
	return pickOne(raw, opusProjectKeys...)
}

// UpdateProject changes the fields passed and nothing else.
func (o Opus) UpdateProject(ctx context.Context, userID, projectID int64, fields OpusProjectFields) ([]byte, error) {
	body := projectBody(fields, false)
	if len(body) == 0 {
		return nil, errors.New("pass at least one of name, icon or slug")
	}
	raw, err := o.gw.Do(ctx, userID, Request{App: AppOpus, Method: http.MethodPut, Path: fmt.Sprintf(opusProjectPath, projectID), Body: body})
	if err != nil {
		return nil, err
	}
	return pickOne(raw, opusProjectKeys...)
}

// DeleteProject deletes a project.
func (o Opus) DeleteProject(ctx context.Context, userID, projectID int64) error {
	_, err := o.gw.Do(ctx, userID, Request{App: AppOpus, Method: http.MethodDelete, Path: fmt.Sprintf(opusProjectPath, projectID), Body: nil})
	return err
}

func projectBody(fields OpusProjectFields, skipEmpty bool) map[string]any {
	body := map[string]any{}
	for key, value := range map[string]*string{"name": fields.Name, "icon": fields.Icon, "slug": fields.Slug} {
		if value != nil && (!skipEmpty || *value != "") {
			body[key] = *value
		}
	}
	return body
}
