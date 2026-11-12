package suite

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestWriteWithoutGrantedScopeExplainsItself(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	w.readOnly = true
	gw := NewGateway(w.tokens(), w.cfg, nil)
	if _, err := gw.Get(context.Background(), 1, AppSablier, "/api/time-entries/running", nil); err != nil {
		t.Fatalf("reads must keep working when writes are not granted: %v", err)
	}
	_, err := gw.Do(context.Background(), 1, Request{App: AppSablier, Method: http.MethodPost, Path: "/api/agents/sessions", Body: map[string]string{}})
	if !errors.Is(err, ErrWriteNotGranted) || !strings.Contains(err.Error(), "sablier.write") || len(w.writes) != 0 {
		t.Fatalf("want a write-not-granted error and no request, got %v, %d writes", err, len(w.writes))
	}
}
