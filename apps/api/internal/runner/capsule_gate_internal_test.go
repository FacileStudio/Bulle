package runner

import (
	"context"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/nacelle"
)

func TestCapsuleCreateAsksOnlyOnceTheRunIsTainted(t *testing.T) {
	sealer, err := nacelle.NewTool(suite.CapsuleCreateName, "seal", func(context.Context, struct{}) (string, error) { return "", nil })
	if err != nil {
		t.Fatal(err)
	}
	g := trustGate(nacelle.UserText("seal my wifi password"))
	g.addWrites(sealWrites([]nacelle.Tool{sealer}))
	if g.asks(suite.CapsuleCreateName, nil) {
		t.Fatal("a clean run must seal unasked")
	}
	g.observe(context.Background(), nacelle.HookEvent{Point: nacelle.AfterToolCall, Call: nacelle.Call{Name: "courrier_read_thread"}})
	if !g.asks(suite.CapsuleCreateName, nil) {
		t.Fatal("after reading mail, capsule_create must ask the owner")
	}
}
