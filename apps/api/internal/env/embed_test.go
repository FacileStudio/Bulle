package env

import (
	"slices"
	"testing"
)

func TestLoadEmbedOrigins(t *testing.T) {
	cases := []struct {
		value   string
		want    []string
		wantErr bool
	}{
		{"", []string{"https://docs.facile.studio", "https://feuilles.facile.studio", "https://diapo.facile.studio", "https://clap.facile.studio"}, false},
		{"https://docs.example.com, http://localhost:5174", []string{"https://docs.example.com", "http://localhost:5174"}, false},
		{"http://docs.example.com", nil, true},
		{"https://docs.example.com/page", nil, true},
		{"docs.example.com", nil, true},
		{"*", nil, true},
	}
	for _, c := range cases {
		t.Setenv("BULLE_EMBED_ORIGINS", c.value)
		got, err := loadEmbedOrigins()
		if (err != nil) != c.wantErr || !slices.Equal(got, c.want) {
			t.Errorf("%q: got %v, err %v", c.value, got, err)
		}
	}
}
