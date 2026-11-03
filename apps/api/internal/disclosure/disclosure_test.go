package disclosure

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FacileStudio/Bulle/core/jev"
)

func screenAnswering(t *testing.T, status int, choice string, confidence float64) *Screen {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(jev.Response{Answers: map[string]jev.Answer{question: {Choice: choice, Confidence: confidence}}})
	}))
	t.Cleanup(server.Close)
	return New(jev.Config{BaseURL: server.URL, APIKey: "test", Attempts: 1})
}

func TestAllows(t *testing.T) {
	cases := []struct {
		name   string
		screen *Screen
		want   bool
	}{
		{"confident shareable", screenAnswering(t, http.StatusOK, optionSafe, 0.9), true},
		{"unsure shareable", screenAnswering(t, http.StatusOK, optionSafe, 0.4), false},
		{"private", screenAnswering(t, http.StatusOK, optionLeak, 0.9), false},
		{"failed call", screenAnswering(t, http.StatusUnauthorized, optionSafe, 0.9), false},
		{"no screen", nil, false},
	}
	for _, c := range cases {
		if got := c.screen.Allows(context.Background(), "hello"); got != c.want {
			t.Errorf("%s: Allows = %v, want %v", c.name, got, c.want)
		}
	}
}
