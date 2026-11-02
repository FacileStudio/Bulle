package suite

import "testing"

func TestFromEnvRequiresPairs(t *testing.T) {
	t.Setenv("SUITE_COURRIER_URL", "https://courrier.test/")
	if _, err := FromEnv("i", "c", "s"); err == nil {
		t.Fatal("a URL without an audience must stop the boot")
	}
	t.Setenv("SUITE_COURRIER_AUDIENCE", "aud")
	cfg, err := FromEnv("i", "c", "s")
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.Enabled(); len(got) != 1 || got[0] != AppCourrier || cfg.Apps[AppCourrier].URL != "https://courrier.test" {
		t.Fatalf("unexpected config %+v", cfg)
	}
}

func TestFromEnvEmptyDisablesEverything(t *testing.T) {
	cfg, err := FromEnv("i", "c", "s")
	if err != nil || len(cfg.Enabled()) != 0 {
		t.Fatalf("want no app enabled, got %+v, %v", cfg.Enabled(), err)
	}
}
