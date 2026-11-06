package runner

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

type fakePolicy struct {
	mu    sync.Mutex
	drops []string
}

func (p *fakePolicy) describe(_ context.Context, input json.RawMessage) (map[string]any, string, error) {
	var in heldInput
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, "", err
	}
	if in.HeldID == "h404" {
		return nil, "", errors.New("no held request h404")
	}
	origin := "https://shop.test"
	if in.HeldID == "h9" {
		origin = "https://other.test"
	}
	return map[string]any{"method": "POST", "url": origin + "/order", "held_id": in.HeldID}, origin, nil
}

func (p *fakePolicy) denied(_ context.Context, input json.RawMessage) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.drops = append(p.drops, string(input))
}

func policyGate(t *testing.T) (*gate, *fakePolicy, store.Store) {
	t.Helper()
	st := store.NewMemory()
	em := &emitter{store: st, hub: newHub(), logger: slog.Default(), runID: "r1"}
	g := newGate(map[string]bool{browserConfirmTool: true}, st, em, "r1", time.Now)
	policy := &fakePolicy{}
	g.policies = map[string]approvalPolicy{browserConfirmTool: policy}
	return g, policy, st
}

func askAndDecide(t *testing.T, g *gate, callID, input string, approve bool) bool {
	t.Helper()
	g.note(callID, browserConfirmTool, input)
	result := make(chan bool, 1)
	go func() { result <- g.approve(context.Background(), browserConfirmTool, json.RawMessage(input)) }()
	for deadline := time.Now().Add(2 * time.Second); !g.pending(callID); {
		if time.Now().After(deadline) {
			t.Fatalf("%s never asked the owner", callID)
		}
		time.Sleep(5 * time.Millisecond)
	}
	if err := g.decide(context.Background(), callID, approve, "1", Edit{}, false); err != nil {
		t.Fatalf("decide %s: %v", callID, err)
	}
	return <-result
}

func interrupts(t *testing.T, st store.Store) []string {
	t.Helper()
	events, err := st.Events(context.Background(), "r1", 0)
	if err != nil {
		t.Fatalf("events: %v", err)
	}
	var out []string
	for _, event := range events {
		if event.Type == EventInterrupt {
			out = append(out, string(event.Payload))
		}
	}
	return out
}

func TestBrowserConfirmShowsTheDaemonRecordAndCoversTheOrigin(t *testing.T) {
	g, _, st := policyGate(t)
	if !askAndDecide(t, g, "c1", `{"held_id":"h1","note":"model text"}`, true) {
		t.Fatal("approved call was refused")
	}
	shown := interrupts(t, st)
	if len(shown) != 1 || !strings.Contains(shown[0], `"url":"https://shop.test/order"`) || strings.Contains(shown[0], "model text") {
		t.Fatalf("interrupt = %v", shown)
	}
	g.note("c2", browserConfirmTool, `{"held_id":"h2"}`)
	if !g.approve(context.Background(), browserConfirmTool, json.RawMessage(`{"held_id":"h2"}`)) {
		t.Fatal("a second request to an approved origin must not ask again")
	}
	if len(interrupts(t, st)) != 1 {
		t.Fatal("the approved origin raised a second approval card")
	}
}

func TestBrowserConfirmDenialDropsTheRequestAndOtherOriginsStillAsk(t *testing.T) {
	g, policy, _ := policyGate(t)
	askAndDecide(t, g, "c1", `{"held_id":"h1"}`, true)
	if askAndDecide(t, g, "c9", `{"held_id":"h9"}`, false) {
		t.Fatal("denied call ran")
	}
	if len(policy.drops) != 1 || !strings.Contains(policy.drops[0], "h9") {
		t.Fatalf("drops = %v", policy.drops)
	}
}

func TestBrowserConfirmOfAnUnknownHoldIsRefusedUnasked(t *testing.T) {
	g, _, st := policyGate(t)
	g.note("c1", browserConfirmTool, `{"held_id":"h404"}`)
	if g.approve(context.Background(), browserConfirmTool, json.RawMessage(`{"held_id":"h404"}`)) {
		t.Fatal("an undescribable call ran")
	}
	if len(interrupts(t, st)) != 0 {
		t.Fatal("an undescribable call reached the owner")
	}
}

func rememberGate(t *testing.T) (*gate, store.Store) {
	t.Helper()
	st := store.NewMemory()
	em := &emitter{store: st, hub: newHub(), logger: slog.Default(), runID: "r1"}
	outward := map[string]bool{"opus_create_task": true, "opus_update_task": true, "opus_delete_task": true, "files_save": true, mailSendTool: true}
	return newGate(outward, st, em, "r1", time.Now), st
}

func askTool(t *testing.T, g *gate, tool, callID string, remember bool) bool {
	t.Helper()
	input := `{"id":"` + callID + `"}`
	g.note(callID, tool, input)
	result := make(chan bool, 1)
	go func() { result <- g.approve(context.Background(), tool, json.RawMessage(input)) }()
	for deadline := time.Now().Add(2 * time.Second); !g.pending(callID); {
		if time.Now().After(deadline) {
			t.Fatalf("%s never asked the owner", callID)
		}
		time.Sleep(5 * time.Millisecond)
	}
	if err := g.decide(context.Background(), callID, true, "1", Edit{}, remember); err != nil {
		t.Fatalf("decide %s: %v", callID, err)
	}
	return <-result
}

func TestRememberCoversLaterCallsOfTheSameTool(t *testing.T) {
	g, st := rememberGate(t)
	if !askTool(t, g, "opus_create_task", "c1", true) {
		t.Fatal("approved call was refused")
	}
	g.note("c2", "opus_create_task", `{"id":"c2"}`)
	if !g.approve(context.Background(), "opus_create_task", json.RawMessage(`{"id":"c2"}`)) {
		t.Fatal("a remembered tool asked again")
	}
	if len(interrupts(t, st)) != 1 {
		t.Fatal("a covered call raised an approval card")
	}
	if g.pending("c2") {
		t.Fatal("a covered call left an approval pending")
	}
}

func TestRememberDoesNotCoverAnotherTool(t *testing.T) {
	g, _ := rememberGate(t)
	askTool(t, g, "opus_create_task", "c1", true)
	if !askTool(t, g, "opus_update_task", "c2", false) {
		t.Fatal("approved call was refused")
	}
}

func TestRememberNeverCoversMail(t *testing.T) {
	g, st := rememberGate(t)
	askTool(t, g, mailSendTool, "c1", true)
	askTool(t, g, mailSendTool, "c2", false)
	if len(interrupts(t, st)) != 2 {
		t.Fatalf("mail was not asked twice: %v", interrupts(t, st))
	}
}

func TestRememberIsIgnoredWithoutApprovalAndDiesWithTheGate(t *testing.T) {
	g, _ := rememberGate(t)
	if g.granted("opus_create_task", "") {
		t.Fatal("a new gate carries a grant")
	}
	g.note("c1", "opus_create_task", `{"id":"c1"}`)
	go g.approve(context.Background(), "opus_create_task", json.RawMessage(`{"id":"c1"}`))
	for deadline := time.Now().Add(2 * time.Second); !g.pending("c1"); {
		if time.Now().After(deadline) {
			t.Fatal("never asked")
		}
		time.Sleep(5 * time.Millisecond)
	}
	if err := g.decide(context.Background(), "c1", false, "1", Edit{}, true); err != nil {
		t.Fatal(err)
	}
	if g.granted("opus_create_task", "") {
		t.Fatal("a denial granted the tool")
	}
	again, _ := rememberGate(t)
	if again.granted("opus_create_task", "") {
		t.Fatal("a new gate carries a grant")
	}
}

func TestRememberNeverCoversDeletesFilesOrURLGatedTools(t *testing.T) {
	g, st := rememberGate(t)
	g.asks = map[string]func(json.RawMessage) bool{"web_fetch": func(json.RawMessage) bool { return true }}
	for _, tool := range []string{"opus_delete_task", "files_save", "web_fetch"} {
		askTool(t, g, tool, tool+"-1", true)
		if g.granted(tool, "") {
			t.Fatalf("%s took a run-wide grant", tool)
		}
	}
	if len(interrupts(t, st)) != 3 {
		t.Fatalf("each gated call must raise its own card: %v", interrupts(t, st))
	}
}
