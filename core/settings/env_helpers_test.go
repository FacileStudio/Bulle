package settings

import "testing"

// clearEnv makes a variable read as unmentioned rather than as empty text.
func clearEnv(t *testing.T, names ...string) {
	t.Helper()
	for _, name := range names {
		t.Setenv(name, "")
	}
}
