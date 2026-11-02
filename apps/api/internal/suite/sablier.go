package suite

import (
	"context"
	"fmt"
	"net/http"
)

const sablierRunningPath = "/api/time-entries/running"

type Sablier struct{ gw *Gateway }

func NewSablier(gw *Gateway) Sablier { return Sablier{gw: gw} }

func (s Sablier) Running(ctx context.Context, userID int64) ([]byte, error) {
	return s.gw.Get(ctx, userID, AppSablier, sablierRunningPath, nil)
}

const (
	sablierProjectsPath = "/api/projects/"
	sablierTasksPath    = "/api/projects/%d/tasks"
	sablierStartPath    = "/api/time-entries/start"
	sablierStopPath     = "/api/time-entries/stop"
)

func (s Sablier) Projects(ctx context.Context, userID, projectID int64) ([]byte, error) {
	if projectID > 0 {
		return s.gw.Get(ctx, userID, AppSablier, fmt.Sprintf(sablierTasksPath, projectID), nil)
	}
	return s.gw.Get(ctx, userID, AppSablier, sablierProjectsPath, nil)
}

func (s Sablier) Start(ctx context.Context, userID, projectID, taskID int64) ([]byte, error) {
	payload := map[string]any{"project_id": projectID, "task_id": taskID}
	return s.gw.Do(ctx, userID, Request{App: AppSablier, Method: http.MethodPost, Path: sablierStartPath, Body: payload})
}

func (s Sablier) Stop(ctx context.Context, userID int64) ([]byte, error) {
	return s.gw.Do(ctx, userID, Request{App: AppSablier, Method: http.MethodPost, Path: sablierStopPath, Body: nil})
}
