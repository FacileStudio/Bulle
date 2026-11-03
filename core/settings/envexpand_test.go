package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFillsPlaceholdersFromTheEnvironment(t *testing.T) {
	t.Setenv("BULLE_TEST_SERVER", "https://bulle.example.org")
	t.Setenv("BULLE_TEST_PROMPT", "line one: x\nline two")
	path := filepath.Join(t.TempDir(), "config.yml")
	content := "session:\n  additional_prompt: \"${BULLE_TEST_PROMPT}\"\n" +
		"server:\n  url: ${BULLE_TEST_SERVER}\n" +
		"sources:\n  skill_dirs: [\"${BULLE_TEST_SERVER}\"]\n" +
		"editor:\n  editor: echo $HOME ${BULLE_TEST_UNSET}\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.URL != "https://bulle.example.org" || len(cfg.SkillDirs) != 1 || cfg.SkillDirs[0] != "https://bulle.example.org" {
		t.Errorf("server %q, skill dirs %v, want the placeholders filled", cfg.Server.URL, cfg.SkillDirs)
	}
	if cfg.Additional != "line one: x\nline two" {
		t.Errorf("additional = %q, want the multi-line value intact", cfg.Additional)
	}
	if cfg.Editor.Editor != "echo $HOME " {
		t.Errorf("editor = %q, want $HOME kept and the unset placeholder emptied", cfg.Editor.Editor)
	}
}
