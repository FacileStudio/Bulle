package agent

import (
	"testing"
)

func TestLocalToolsFiltersDisabledSearchTools(t *testing.T) {
	dir := t.TempDir()
	off := false

	config := Config{
		Session: Session{Root: dir},
		Toggles: Toggles{
			Bash:          &off,
			FindFiles:     &off,
			SearchContent: &off,
		},
		Security: Security{PathIsolation: &off, DenyElevation: &off, EnvIsolation: &off},
	}

	set, local, err := localTools(config)
	if err != nil {
		t.Fatalf("localTools: %v", err)
	}
	defer set.Close()

	for _, tool := range local {
		if tool.Name() == "glob" {
			t.Error("glob should be filtered when FindFiles toggle is false")
		}
		if tool.Name() == "grep" {
			t.Error("grep should be filtered when SearchContent toggle is false")
		}
	}
}

func TestLocalToolsMountsSearchToolsByDefault(t *testing.T) {
	dir := t.TempDir()
	off := false

	config := Config{
		Session: Session{Root: dir},
		Toggles: Toggles{
			Bash: &off,
		},
		Security: Security{PathIsolation: &off, DenyElevation: &off, EnvIsolation: &off},
	}

	set, local, err := localTools(config)
	if err != nil {
		t.Fatalf("localTools: %v", err)
	}
	defer set.Close()

	hasFind, hasSearch := false, false
	for _, tool := range local {
		if tool.Name() == "glob" {
			hasFind = true
		}
		if tool.Name() == "grep" {
			hasSearch = true
		}
	}
	if !hasFind {
		t.Error("glob missing when FindFiles toggle is default")
	}
	if !hasSearch {
		t.Error("grep missing when SearchContent toggle is default")
	}
}
