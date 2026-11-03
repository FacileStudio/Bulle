package settings

import (
	"os"
	"testing"
)

func TestLegacyNacelleEnvStillResolves(t *testing.T) {
	t.Setenv("NACELLE_ROOT", "legacy")
	t.Setenv("BULLE_ROOT", "current")
	if got := envGet("ROOT"); got != "current" {
		t.Fatalf("BULLE_ must win over NACELLE_: got %q", got)
	}

	t.Setenv("BULLE_ROOT", "")
	if err := os.Unsetenv("BULLE_ROOT"); err != nil {
		t.Fatal(err)
	}
	if got := envGet("ROOT"); got != "legacy" {
		t.Fatalf("NACELLE_ROOT must still resolve: got %q", got)
	}
}
