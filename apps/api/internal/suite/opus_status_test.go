package suite

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestOpusStatusMatchesColumns(t *testing.T) {
	var columns []OpusColumn
	if err := json.Unmarshal([]byte(opusColumnsBody), &columns); err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"to-do": "to-do", "todo": "to-do", "To Do": "to-do", "in_progress": "in-progress", "In Progress": "in-progress",
		"IN-REVIEW": "in-review", "done": "done", "completed": "done", "closed": "done", "planned": "planned", "archived": "archived",
	}
	for in, want := range cases {
		if got, ok := matchStatus(columns, in); !ok || got != want {
			t.Errorf("status %q: want %q, got %q (%t)", in, want, got, ok)
		}
	}
	for _, miss := range []string{"", "bogus", "progress"} {
		if got, ok := matchStatus(columns, miss); ok {
			t.Errorf("status %q must not match, got %q", miss, got)
		}
	}
}

func TestOpusStatusMissListsValidStatuses(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	opus := NewOpus(NewGateway(w.tokens(), w.cfg, nil))
	for _, query := range []OpusQuery{{Status: "bogus"}, {Project: "p", Status: "bogus"}} {
		tasks, err := opus.Tasks(context.Background(), 1, query)
		if err == nil || !strings.Contains(err.Error(), "to-do, in-progress, in-review, done, planned, archived") {
			t.Fatalf("a status that matches nothing must list the valid ones, got %v, %v", tasks, err)
		}
	}
	title, status := "t", "bogus"
	if _, err := opus.Create(context.Background(), 1, "", "p", OpusTaskFields{Title: &title, Status: &status}); err == nil || len(w.writes) != 0 {
		t.Fatalf("a create with an unknown status must fail before any write: %v, %+v", err, w.writes)
	}
}

func TestOpusWritesSendTheResolvedSlug(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	opus := NewOpus(NewGateway(w.tokens(), w.cfg, nil))
	title, status := "t", "In_Progress"
	if _, err := opus.Create(context.Background(), 1, "", "p", OpusTaskFields{Title: &title, Status: &status}); err != nil {
		t.Fatal(err)
	}
	closed := "closed"
	if _, err := opus.Update(context.Background(), 1, 9, "", OpusTaskFields{Status: &closed}); err != nil {
		t.Fatal(err)
	}
	if len(w.writes) != 2 || !strings.Contains(w.writes[0].body, `"status":"in-progress"`) || !strings.Contains(w.writes[1].body, `"status":"done"`) {
		t.Fatalf("writes must carry the column slug: %+v", w.writes)
	}
	mine, err := opus.Tasks(context.Background(), 1, OpusQuery{Status: "todo"})
	if err != nil || len(mine) != 1 || mine[0].Status != "to-do" {
		t.Fatalf("the my-tasks filter must resolve todo to to-do: %+v, %v", mine, err)
	}
}

func TestOpusProjectListsColumns(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	raw, err := NewOpus(NewGateway(w.tokens(), w.cfg, nil)).Project(context.Background(), 1, "", 5)
	if err != nil || !strings.Contains(string(raw), `"columns":[{"slug":"to-do"`) || !strings.Contains(string(raw), `"isFinal":true`) {
		t.Fatalf("the project must carry its columns: %s, %v", raw, err)
	}
}
