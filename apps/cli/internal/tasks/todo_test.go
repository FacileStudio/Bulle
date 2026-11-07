package tasks

import (
	"encoding/json"
	"testing"
)

func TestTodoWriteReadsTodosAndTheOldTasksShape(t *testing.T) {
	for _, input := range []string{
		`{"todos":[{"content":"build","status":"in_progress"}]}`,
		`{"tasks":[{"title":"build","status":"in_progress"}]}`,
	} {
		plan, err := tasksTool{}.merge(json.RawMessage(input))
		if err != nil || len(plan) != 1 || plan[0].Title != "build" || plan[0].Status != statusActive {
			t.Fatalf("%s: plan %+v err %v", input, plan, err)
		}
	}
	if (tasksTool{}).Name() != "todo_write" {
		t.Fatalf("the checklist tool is %s, want todo_write", tasksTool{}.Name())
	}
}
