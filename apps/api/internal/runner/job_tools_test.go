package runner

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

func findTool(tools []nacelle.Tool, name string) nacelle.Tool {
	for _, t := range tools {
		if t.Name() == name {
			return t
		}
	}
	return nil
}

func TestBuildJobToolsGuard(t *testing.T) {
	st := store.NewMemory()
	if tools := buildJobTools(st, 1, "", store.ChannelCron); tools != nil {
		t.Fatalf("expected nil tools for cron channel, got %d", len(tools))
	}
	if tools := buildJobTools(nil, 1, "", store.ChannelWeb); tools != nil {
		t.Fatalf("expected nil tools for nil store, got %d", len(tools))
	}
	tools := buildJobTools(st, 1, "", store.ChannelWeb)
	if len(tools) != 4 {
		t.Fatalf("expected 4 tools, got %d", len(tools))
	}
	expected := []string{"create_background_job", "list_background_jobs", "update_background_job", "delete_background_job"}
	for _, name := range expected {
		if findTool(tools, name) == nil {
			t.Fatalf("expected tool %s not found", name)
		}
	}
}

func TestCreateBackgroundJobValidation(t *testing.T) {
	st := store.NewMemory()
	tool := findTool(buildJobTools(st, 1, "", store.ChannelWeb), "create_background_job")

	badCases := []string{
		`{"name":"","schedule":"0 9 * * *","prompt":"do work"}`,
		`{"name":"job","schedule":"","prompt":"do work"}`,
		`{"name":"job","schedule":"not-a-cron","prompt":"do work"}`,
		`{"name":"job","schedule":"0 9 * * *","prompt":""}`,
	}
	ctx := context.Background()
	for _, payload := range badCases {
		_, err := tool.Run(ctx, json.RawMessage(payload))
		if err == nil {
			t.Fatalf("expected error for payload: %s", payload)
		}
	}
}

func TestCreateBackgroundJobSuccess(t *testing.T) {
	st := store.NewMemory()
	tool := findTool(buildJobTools(st, 42, "", store.ChannelWeb), "create_background_job")
	ctx := context.Background()

	out, err := tool.Run(ctx, json.RawMessage(`{"name":"morning-digest","schedule":"0 9 * * *","prompt":"check reminders"}`))
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	var created store.Job
	if err := json.Unmarshal([]byte(out), &created); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if created.UserID != 42 || created.Name != "morning-digest" || created.Delivery != "antenne" || !created.Enabled {
		t.Fatalf("unexpected created job: %+v", created)
	}

	stored, err := st.GetJob(ctx, 42, created.ID)
	if err != nil {
		t.Fatalf("get job from store: %v", err)
	}
	if stored.Name != "morning-digest" || stored.Schedule != "0 9 * * *" {
		t.Fatalf("unexpected stored job: %+v", stored)
	}
}

func TestCreateBackgroundJobCustomDelivery(t *testing.T) {
	st := store.NewMemory()
	tool := findTool(buildJobTools(st, 7, "", store.ChannelWeb), "create_background_job")
	ctx := context.Background()

	out, err := tool.Run(ctx, json.RawMessage(`{"name":"alert","schedule":"*/5 * * * *","prompt":"ping","delivery":"matrix"}`))
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	var created store.Job
	if err := json.Unmarshal([]byte(out), &created); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if created.Delivery != "matrix" {
		t.Fatalf("expected delivery matrix, got %s", created.Delivery)
	}
}

func TestListBackgroundJobs(t *testing.T) {
	st := store.NewMemory()
	ctx := context.Background()
	toolUser1 := findTool(buildJobTools(st, 1, "", store.ChannelWeb), "list_background_jobs")

	out, err := toolUser1.Run(ctx, json.RawMessage(`{}`))
	if err != nil || out != "[]" {
		t.Fatalf("expected empty list, got err=%v out=%s", err, out)
	}

	if _, err := st.CreateJob(ctx, store.Job{UserID: 1, Name: "job1", Schedule: "0 9 * * *", Prompt: "p1", Enabled: true}); err != nil {
		t.Fatalf("create job1: %v", err)
	}
	if _, err := st.CreateJob(ctx, store.Job{UserID: 1, Name: "job2", Schedule: "0 10 * * *", Prompt: "p2", Enabled: true}); err != nil {
		t.Fatalf("create job2: %v", err)
	}
	if _, err := st.CreateJob(ctx, store.Job{UserID: 2, Name: "job3", Schedule: "0 11 * * *", Prompt: "p3", Enabled: true}); err != nil {
		t.Fatalf("create job3: %v", err)
	}

	out, err = toolUser1.Run(ctx, json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	var list []store.Job
	if err := json.Unmarshal([]byte(out), &list); err != nil || len(list) != 2 {
		t.Fatalf("expected 2 jobs for user 1, got err=%v len=%d", err, len(list))
	}
}

func TestUpdateBackgroundJobValidationAndMissing(t *testing.T) {
	st := store.NewMemory()
	ctx := context.Background()
	tool := findTool(buildJobTools(st, 1, "", store.ChannelWeb), "update_background_job")

	if _, err := tool.Run(ctx, json.RawMessage(`{"id":""}`)); err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, err := tool.Run(ctx, json.RawMessage(`{"id":"non-existent"}`)); err == nil {
		t.Fatal("expected error for non-existent job")
	}

	job, err := st.CreateJob(ctx, store.Job{UserID: 1, Name: "orig", Schedule: "0 9 * * *", Prompt: "p", Enabled: true})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	badPayload := `{"id":"` + job.ID + `","schedule":"invalid"}`
	if _, err := tool.Run(ctx, json.RawMessage(badPayload)); err == nil {
		t.Fatal("expected error for invalid schedule update")
	}
}

func TestUpdateBackgroundJobSuccess(t *testing.T) {
	st := store.NewMemory()
	ctx := context.Background()
	tool := findTool(buildJobTools(st, 1, "", store.ChannelWeb), "update_background_job")

	job, err := st.CreateJob(ctx, store.Job{UserID: 1, Name: "orig", Schedule: "0 9 * * *", Prompt: "p", Enabled: true})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	payload := `{"id":"` + job.ID + `","enabled":false,"schedule":"0 12 * * *","prompt":"new-p","delivery":"matrix"}`
	out, err := tool.Run(ctx, json.RawMessage(payload))
	if err != nil {
		t.Fatalf("update job: %v", err)
	}
	var updated store.Job
	if err := json.Unmarshal([]byte(out), &updated); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if updated.Enabled || updated.Schedule != "0 12 * * *" || updated.Prompt != "new-p" || updated.Delivery != "matrix" {
		t.Fatalf("unexpected updated job: %+v", updated)
	}
	got, err := st.GetJob(ctx, 1, job.ID)
	if err != nil || got.Enabled || got.Schedule != "0 12 * * *" {
		t.Fatalf("store not updated: err=%v got=%+v", err, got)
	}
}

func TestDeleteBackgroundJob(t *testing.T) {
	st := store.NewMemory()
	ctx := context.Background()
	job, err := st.CreateJob(ctx, store.Job{UserID: 1, Name: "to-delete", Schedule: "0 9 * * *", Prompt: "p", Enabled: true})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}

	strangerTool := findTool(buildJobTools(st, 999, "", store.ChannelWeb), "delete_background_job")
	if _, err := strangerTool.Run(ctx, json.RawMessage(`{"id":"`+job.ID+`"}`)); err == nil {
		t.Fatal("stranger should not be able to delete job")
	}

	ownerTool := findTool(buildJobTools(st, 1, "", store.ChannelWeb), "delete_background_job")
	out, err := ownerTool.Run(ctx, json.RawMessage(`{"id":"`+job.ID+`"}`))
	if err != nil || !strings.Contains(out, "deleted") {
		t.Fatalf("delete job: %v out=%s", err, out)
	}

	if _, err := st.GetJob(ctx, 1, job.ID); err == nil {
		t.Fatal("expected job to be deleted from store")
	}
}
