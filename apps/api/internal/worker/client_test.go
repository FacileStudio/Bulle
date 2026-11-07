package worker

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/FacileStudio/tronc/errors"
)

func stub(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	dir, err := os.MkdirTemp("", "bw")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	socket := filepath.Join(dir, "w.sock")
	listener, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{Handler: handler}
	go server.Serve(listener)
	t.Cleanup(func() { server.Close() })
	return New(socket, "secret")
}

func TestCreateSendsTokenAndDecodesTheBox(t *testing.T) {
	client := stub(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret" || r.Method != http.MethodPost || r.URL.Path != "/boxes" {
			t.Errorf("unexpected request %s %s auth=%q", r.Method, r.URL.Path, r.Header.Get("Authorization"))
		}
		var in CreateRequest
		_ = json.NewDecoder(r.Body).Decode(&in)
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(Box{ID: "b1", Name: in.Name, Owner: in.Owner, Status: StatusCreating})
	})
	box, err := client.Create(context.Background(), CreateRequest{Name: "bulle-1", Owner: "1", CPUs: 2})
	if err != nil {
		t.Fatal(err)
	}
	if box.ID != "b1" || box.Owner != "1" || box.Status != StatusCreating {
		t.Fatalf("box = %+v", box)
	}
}

func TestListFiltersByOwner(t *testing.T) {
	client := stub(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("owner") != "9" {
			t.Errorf("owner = %q", r.URL.Query().Get("owner"))
		}
		_ = json.NewEncoder(w).Encode([]Box{{ID: "b9"}})
	})
	boxes, err := client.List(context.Background(), "9")
	if err != nil || len(boxes) != 1 || boxes[0].ID != "b9" {
		t.Fatalf("boxes = %v err = %v", boxes, err)
	}
}

func TestErrorsMapToTheSuiteEnvelope(t *testing.T) {
	cases := map[string]string{
		CodeInvalid:      "invalid_argument",
		CodeNotFound:     "not_found",
		CodeExists:       "already_exists",
		CodeNotStopped:   "failed_precondition",
		CodeBusy:         "unavailable",
		CodeUnauthorized: "internal",
		CodeInternal:     "internal",
	}
	for code, want := range cases {
		client := stub(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "nope", "code": code})
		})
		err := client.Stop(context.Background(), "b1")
		var envelope *errors.Error
		if !asEnvelope(err, &envelope) || envelope.Code != want {
			t.Errorf("%s: err = %v, want code %s", code, err, want)
		}
		if !IsCode(err, code) {
			t.Errorf("%s: IsCode lost the worker code", code)
		}
	}
}

func asEnvelope(err error, target **errors.Error) bool {
	for err != nil {
		if e, ok := err.(*errors.Error); ok {
			*target = e
			return true
		}
		unwrapper, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = unwrapper.Unwrap()
	}
	return false
}

func TestUnreachableSocketIsUnavailable(t *testing.T) {
	client := New(filepath.Join(t.TempDir(), "missing.sock"), "x")
	_, err := client.Get(context.Background(), "b1")
	if errors.Status(err) != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, err = %v", errors.Status(err), err)
	}
}

func TestActionPaths(t *testing.T) {
	var seen []string
	client := stub(t, func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.Path)
		if r.URL.Path == "/boxes/b1/snapshots" {
			_ = json.NewEncoder(w).Encode([]Snapshot{{Tag: "v1"}})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	ctx := context.Background()
	_ = client.Start(ctx, "b1")
	_ = client.Remove(ctx, "b1")
	_ = client.SetSSHKey(ctx, "b1", "k")
	_ = client.SetNet(ctx, "b1", "offline")
	_ = client.Snapshot(ctx, "b1", "v1")
	_ = client.Rollback(ctx, "b1", "v1")
	snapshots, err := client.Snapshots(ctx, "b1")
	if err != nil || len(snapshots) != 1 || snapshots[0].Tag != "v1" {
		t.Fatalf("snapshots = %v err = %v", snapshots, err)
	}
	want := []string{
		"POST /boxes/b1/start", "POST /boxes/b1/rm", "PUT /boxes/b1/ssh-key", "PUT /boxes/b1/net",
		"POST /boxes/b1/snapshot", "POST /boxes/b1/rollback", "GET /boxes/b1/snapshots",
	}
	for i, line := range want {
		if i >= len(seen) || seen[i] != line {
			t.Fatalf("seen = %v, want %v", seen, want)
		}
	}
}
