package suite

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestWriteWithoutGrantedScopeExplainsItself(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	w.readOnly = true
	gw := NewGateway(w.tokens(), w.cfg, nil)
	if _, err := NewSablier(gw).Running(context.Background(), 1); err != nil {
		t.Fatalf("reads must keep working when writes are not granted: %v", err)
	}
	_, err := NewSablier(gw).Stop(context.Background(), 1)
	if !errors.Is(err, ErrWriteNotGranted) || !strings.Contains(err.Error(), "sablier.write") || len(w.writes) != 0 {
		t.Fatalf("want a write-not-granted error and no request, got %v, %d writes", err, len(w.writes))
	}
}
