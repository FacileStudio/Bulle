package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFillsPlaceholdersFromTheEnvironment(t *testing.T) {
	t.Setenv("BULLE_TEST_USER", "@me:example.org")
	t.Setenv("BULLE_TEST_PROMPT", "line one: x\nline two")
	path := filepath.Join(t.TempDir(), "config.yml")
	content := "session:\n  additional_prompt: \"${BULLE_TEST_PROMPT}\"\n" +
		"chat:\n  matrix:\n    user_id: ${BULLE_TEST_USER}\n    allow: [\"${BULLE_TEST_USER}\"]\n    password_command: echo $HOME ${BULLE_TEST_UNSET}\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Chat.Matrix.UserID != "@me:example.org" || len(cfg.Chat.Matrix.Allow) != 1 || cfg.Chat.Matrix.Allow[0] != "@me:example.org" {
		t.Errorf("matrix = %+v, want the placeholders filled", cfg.Chat.Matrix)
	}
	if cfg.Additional != "line one: x\nline two" {
		t.Errorf("additional = %q, want the multi-line value intact", cfg.Additional)
	}
	if cfg.Chat.Matrix.PasswordCommand != "echo $HOME " {
		t.Errorf("password_command = %q, want $HOME kept and the unset placeholder emptied", cfg.Chat.Matrix.PasswordCommand)
	}
}
