package env

import "testing"

func TestLoadBoiteIdleMinutes(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		want    int
		wantErr bool
	}{
		{"default", "", 120, false},
		{"explicit", "45", 45, false},
		{"never", "-1", -1, false},
		{"zero", "0", 0, true},
		{"garbage", "soon", 0, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv("BULLE_IDLE_MINUTES", c.value)
			boite, err := loadBoite()
			if c.wantErr {
				if err == nil {
					t.Fatalf("accepted %q", c.value)
				}
				return
			}
			if err != nil || boite.IdleMinutes != c.want {
				t.Fatalf("IdleMinutes = %d, err = %v", boite.IdleMinutes, err)
			}
		})
	}
}

func TestValidatePushURL(t *testing.T) {
	for value, ok := range map[string]bool{
		"": true, "https://ntfy.example.com": true, "http://10.0.0.1:8080": true,
		"ntfy.example.com": false, "ftp://x": false, "https://": false,
	} {
		if err := validatePushURL(value); (err == nil) != ok {
			t.Fatalf("%q: err = %v", value, err)
		}
	}
}

func TestNativeOriginsAreAccepted(t *testing.T) {
	if err := validateOrigins([]string{"capacitor://localhost", "tauri://localhost", "http://tauri.localhost", "https://localhost", "http://localhost"}); err != nil {
		t.Fatal(err)
	}
	if err := validateOrigins([]string{"capacitor://evil.example"}); err == nil {
		t.Fatal("accepted an unknown capacitor origin")
	}
}

func TestPorteBearerSettingsComeFromEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://bulle@localhost/bulle")
	t.Setenv("OIDC_ISSUER", "https://sso.example.test")
	t.Setenv("OIDC_CLIENT_ID", "bulle")
	t.Setenv("OIDC_CLIENT_SECRET", "test")
	t.Setenv("OIDC_REDIRECT_URL", "https://bulle.example.test/api/auth/oidc/callback")
	t.Setenv("OIDC_SUCCESS_URL", "https://bulle.example.test/")
	t.Setenv("OIDC_MACHINE_AUDIENCE", "bulle")
	t.Setenv("OIDC_SCOPE_PREFIX", "bulle")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Porte.MachineAudience != "bulle" || cfg.Porte.ScopePrefix != "bulle" {
		t.Fatalf("porte bearer settings = %q, %q", cfg.Porte.MachineAudience, cfg.Porte.ScopePrefix)
	}
}

func TestValidateGatewayURL(t *testing.T) {
	cases := map[string]bool{
		"": true,
		"wss://bulle.facile.studio/api/hardware/ws":   true,
		"ws://localhost:4000/api/hardware/ws":         true,
		"https://bulle.facile.studio/api/hardware/ws": false,
		"wss:///api/hardware/ws":                      false,
	}
	for value, ok := range cases {
		if err := validateGatewayURL(value); (err == nil) != ok {
			t.Errorf("validateGatewayURL(%q) = %v, want ok=%v", value, err, ok)
		}
	}
}
