package settings

import (
	"testing"
)

func TestSearchToolsDefaultOn(t *testing.T) {
	written(t, "")

	config, err := settings(Config{})
	if err != nil {
		t.Fatalf("settings: %v", err)
	}
	if !*config.SearchContent {
		t.Error("grep default off")
	}
	if !*config.FindFiles {
		t.Error("glob default off")
	}
}

func TestSearchToolsCanBeDisabledByFile(t *testing.T) {
	written(t, "tools:\n  grep: false\n  glob: false\n")

	config, err := settings(Config{})
	if err != nil {
		t.Fatalf("settings: %v", err)
	}
	if *config.SearchContent {
		t.Error("grep: false in file left toggle on")
	}
	if *config.FindFiles {
		t.Error("glob: false in file left toggle on")
	}
}

func TestSearchToolsEnvironmentBeatsFile(t *testing.T) {
	written(t, "tools:\n  grep: false\n  glob: false\n")
	t.Setenv("BULLE_SEARCH_CONTENT", "true")
	t.Setenv("BULLE_FIND_FILES", "true")

	config, err := settings(Config{})
	if err != nil {
		t.Fatalf("settings: %v", err)
	}
	if !*config.SearchContent {
		t.Error("BULLE_SEARCH_CONTENT environment did not beat file")
	}
	if !*config.FindFiles {
		t.Error("BULLE_FIND_FILES environment did not beat file")
	}
}

func TestPreV032ToolKeysStillRead(t *testing.T) {
	written(t, "tools:\n  run_command: true\n  tasks: false\n  search_content: false\n  find_files: false\n  glob: true\n")

	config, err := settings(Config{})
	if err != nil {
		t.Fatalf("settings: %v", err)
	}
	if !*config.Bash || *config.Tasks || *config.SearchContent || !*config.FindFiles {
		t.Errorf("old keys must alias the new ones, new keys winning: bash=%v todo_write=%v grep=%v glob=%v",
			*config.Bash, *config.Tasks, *config.SearchContent, *config.FindFiles)
	}
}
