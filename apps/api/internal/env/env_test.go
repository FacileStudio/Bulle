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
