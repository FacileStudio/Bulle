package suite

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestWriteWithoutGrantedScopeExplainsItself(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	w.readOnly = true
	gw := NewGateway(w.tokens(), w.cfg, nil)
	if _, err := NewSablier(gw).Running(context.Background(), 1); err != nil {
		t.Fatalf("reads must keep working when writes are not granted: %v", err)
	}
	_, err := NewSablier(gw).Stop(context.Background(), 1)
	if !errors.Is(err, ErrWriteNotGranted) || !strings.Contains(err.Error(), "sablier.write") || len(w.writes) != 0 {
		t.Fatalf("want a write-not-granted error and no request, got %v, %d writes", err, len(w.writes))
	}
}

func TestOpusCreateProjectSendsWorkspaceAndName(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	name := "New Project"
	_, err := NewOpus(NewGateway(w.tokens(), w.cfg, nil)).CreateProject(context.Background(), 1, "", OpusProjectFields{Name: &name})
	if err != nil || len(w.writes) != 1 || w.writes[0].method != "POST" || w.writes[0].path != "/api/project" {
		t.Fatalf("want one POST to /api/project, got %+v, %v", w.writes, err)
	}
	var sent map[string]any
	_ = json.Unmarshal([]byte(w.writes[0].body), &sent)
	if sent["name"] != "New Project" || sent["workspaceId"] != float64(2) {
		t.Fatalf("create must send the name and resolved workspace id: %v", sent)
	}
}

func TestOpusUpdateProjectOnlySendsPassedFields(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	name := "Renamed"
	_, err := NewOpus(NewGateway(w.tokens(), w.cfg, nil)).UpdateProject(context.Background(), 1, 7, OpusProjectFields{Name: &name})
	if err != nil || len(w.writes) != 1 || w.writes[0].method != "PUT" || w.writes[0].path != "/api/project/7" {
		t.Fatalf("want one PUT to /api/project/7, got %+v, %v", w.writes, err)
	}
	var sent map[string]any
	_ = json.Unmarshal([]byte(w.writes[0].body), &sent)
	if sent["name"] != "Renamed" {
		t.Fatalf("update must send only the passed fields: %v", sent)
	}
	if _, leaked := sent["id"]; leaked {
		t.Fatalf("read-only fields must not be sent back: %v", sent)
	}
}

func TestOpusUpdateProjectNeedsAField(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	if _, err := NewOpus(NewGateway(w.tokens(), w.cfg, nil)).UpdateProject(context.Background(), 1, 7, OpusProjectFields{}); err == nil || len(w.writes) != 0 {
		t.Fatalf("an empty update must be refused before any write: %v, %+v", err, w.writes)
	}
}

func TestOpusUpdateProjectSchemaDoesNotRequireName(t *testing.T) {
	field, _ := reflect.TypeOf(OpusProjectFields{}).FieldByName("Name")
	if strings.HasPrefix(field.Tag.Get("jsonschema"), "required") {
		t.Fatal("name is required only on create; update takes any one field")
	}
}

func TestOpusProjectResolvesOutsideTheDefaultWorkspace(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	opus := NewOpus(NewGateway(w.tokens(), w.cfg, nil))
	if _, err := opus.Project(context.Background(), 1, "", 8); err == nil {
		t.Fatal("a project of another workspace must not resolve in the default workspace")
	}
	if _, err := opus.Project(context.Background(), 1, "Other", 8); err != nil {
		t.Fatalf("the project must resolve once its workspace is passed: %v", err)
	}
	if _, ok := reflect.TypeOf(opusShowProjectInput{}).FieldByName("Workspace"); !ok {
		t.Fatal("opus_get_project must accept a workspace")
	}
	if _, ok := reflect.TypeOf(opusProjectInput{}).FieldByName("Workspace"); ok {
		t.Fatal("opus_delete_project keeps the id-only input")
	}
}
