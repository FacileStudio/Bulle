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
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
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
	outward := map[string]bool{
		"opus_create_task": true, "opus_update_task": true, "opus_delete_task": true,
		"opus_delete_project": true, "agenda_delete_event": true, "courrier_delete_draft": true,
		"courrier_update": true, "files_save": true, mailSendTool: true,
	}
	g := newGate(outward, st, em, "r1", time.Now)
	defaults := suite.Defaults{CourrierAccount: 5, AgendaCalendar: 7, OpusWorkspace: "ws1"}
	space := store.Space{Name: "Work"}
	for _, name := range []string{
		"opus_create_task", "opus_update_task", "opus_delete_task", "opus_delete_project",
		"agenda_delete_event", "courrier_delete_draft", "courrier_update", "files_save",
	} {
		g.addPolicy(name, suitePolicy{name: name, defaults: defaults, space: space, gate: g})
	}
	return g, st
}

func askTool(t *testing.T, g *gate, tool, callID string, remember bool) bool {
	t.Helper()
	return scopeAsk(t, g, tool, callID, `{"id":"`+callID+`"}`, remember)
}

func scopeAsk(t *testing.T, g *gate, tool, callID, input string, remember bool) bool {
	t.Helper()
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
	if g.granted("opus_create_task", "workspace:ws1") {
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
	if g.granted("opus_create_task", "workspace:ws1") {
		t.Fatal("a denial granted the tool")
	}
	again, _ := rememberGate(t)
	if again.granted("opus_create_task", "workspace:ws1") {
		t.Fatal("a new gate carries a grant")
	}
}

func TestRememberNeverCoversDeletesFilesOrURLGatedTools(t *testing.T) {
	g, st := rememberGate(t)
	g.asks = map[string]func(json.RawMessage) bool{"web_fetch": func(json.RawMessage) bool { return true }}
	for _, tool := range []string{"opus_delete_task", "files_save", "web_fetch"} {
		askTool(t, g, tool, tool+"-1", true)
	}
	if len(g.grants) != 0 {
		t.Fatal("a gated tool took a grant")
	}
	if len(interrupts(t, st)) != 3 {
		t.Fatalf("each gated call must raise its own card: %v", interrupts(t, st))
	}
}

func TestRememberReleasesSameToolCallsAlreadyWaiting(t *testing.T) {
	g, _ := rememberGate(t)
	results := map[string]chan bool{}
	for _, id := range []string{"w1", "w2", "w3"} {
		input := `{"id":"` + id + `"}`
		g.note(id, "opus_create_task", input)
		ch := make(chan bool, 1)
		results[id] = ch
		go func() { ch <- g.approve(context.Background(), "opus_create_task", json.RawMessage(input)) }()
	}
	for _, id := range []string{"w1", "w2", "w3"} {
		for deadline := time.Now().Add(2 * time.Second); !g.pending(id); {
			if time.Now().After(deadline) {
				t.Fatalf("%s never asked", id)
			}
			time.Sleep(5 * time.Millisecond)
		}
	}
	if err := g.decide(context.Background(), "w1", true, "1", Edit{}, true); err != nil {
		t.Fatal(err)
	}
	for id, ch := range results {
		select {
		case ok := <-ch:
			if !ok {
				t.Fatalf("%s was refused", id)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("%s stayed waiting after the grant", id)
		}
	}
}

func TestRememberedWriteCoversOnlyTheSameTarget(t *testing.T) {
	g, st := rememberGate(t)
	scopeAsk(t, g, "opus_update_task", "c1", `{"id":1}`, true)
	scopeAsk(t, g, "opus_update_task", "c2", `{"id":2}`, false)
	if len(interrupts(t, st)) != 2 {
		t.Fatalf("a remembered task 1 covered task 2: %v", interrupts(t, st))
	}
	g.note("c3", "opus_update_task", `{"id":1}`)
	if !g.approve(context.Background(), "opus_update_task", json.RawMessage(`{"id":1}`)) {
		t.Fatal("a second update of the approved task asked again")
	}
	if len(interrupts(t, st)) != 2 || g.pending("c3") {
		t.Fatalf("the approved target raised another card: %v", interrupts(t, st))
	}
}

func TestRememberedWriteDoesNotReleaseAnotherTarget(t *testing.T) {
	g, _ := rememberGate(t)
	inputs := map[string]string{"w1": `{"id":1}`, "w2": `{"id":2}`}
	results := map[string]chan bool{}
	for _, id := range []string{"w1", "w2"} {
		input := inputs[id]
		g.note(id, "opus_update_task", input)
		ch := make(chan bool, 1)
		results[id] = ch
		go func() { ch <- g.approve(context.Background(), "opus_update_task", json.RawMessage(input)) }()
	}
	for _, id := range []string{"w1", "w2"} {
		for deadline := time.Now().Add(2 * time.Second); !g.pending(id); {
			if time.Now().After(deadline) {
				t.Fatalf("%s never asked", id)
			}
			time.Sleep(5 * time.Millisecond)
		}
	}
	if err := g.decide(context.Background(), "w1", true, "1", Edit{}, true); err != nil {
		t.Fatal(err)
	}
	if !g.pending("w2") {
		t.Fatal("a remembered write released a waiting call on another task")
	}
	select {
	case ok := <-results["w2"]:
		t.Fatalf("the other target was released approved=%v", ok)
	default:
	}
	if err := g.decide(context.Background(), "w2", true, "1", Edit{}, false); err != nil {
		t.Fatal(err)
	}
	if ok := <-results["w2"]; !ok {
		t.Fatal("the waiting call was refused")
	}
}

func TestRememberedCourrierActionDoesNotCoverAnotherAction(t *testing.T) {
	g, st := rememberGate(t)
	scopeAsk(t, g, "courrier_update", "c1", `{"ids":[1],"action":"mark_read"}`, true)
	scopeAsk(t, g, "courrier_update", "c2", `{"ids":[1],"action":"delete"}`, false)
	if len(interrupts(t, st)) != 2 {
		t.Fatalf("a remembered mark_read covered a delete: %v", interrupts(t, st))
	}
	g.note("c3", "courrier_update", `{"ids":[1],"action":"mark_read"}`)
	if !g.approve(context.Background(), "courrier_update", json.RawMessage(`{"ids":[1],"action":"mark_read"}`)) {
		t.Fatal("a second mark_read on the approved account asked again")
	}
	if len(interrupts(t, st)) != 2 || g.pending("c3") {
		t.Fatalf("the approved action raised another card: %v", interrupts(t, st))
	}
}

func TestDeletesAndFilesNeverTakeAGrant(t *testing.T) {
	g, st := rememberGate(t)
	cases := []struct {
		tool  string
		input string
	}{
		{"opus_delete_task", `{"id":1}`},
		{"opus_delete_project", `{"id":1}`},
		{"agenda_delete_event", `{"id":1}`},
		{"courrier_delete_draft", `{"id":1}`},
		{"files_save", `{"content":"x","name":"n.txt"}`},
	}
	for i, tc := range cases {
		scopeAsk(t, g, tc.tool, "d"+string(rune('0'+i)), tc.input, true)
		scopeAsk(t, g, tc.tool, "d"+string(rune('0'+i))+"-again", tc.input, false)
	}
	if len(interrupts(t, st)) != 2*len(cases) {
		t.Fatalf("a gated write was not asked every time: %v", interrupts(t, st))
	}
	if len(g.grants) != 0 {
		t.Fatalf("a gated write recorded a grant: %v", g.grants)
	}
	if g.granted("opus_delete_task", "task:1") || g.granted("files_save", "folder:0") {
		t.Fatal("a gated write is grantable by scope")
	}
}

func TestCourrierDeleteActionNeverTakesAGrant(t *testing.T) {
	g, st := rememberGate(t)
	input := `{"ids":[1],"action":"delete"}`
	scopeAsk(t, g, "courrier_update", "c1", input, true)
	scopeAsk(t, g, "courrier_update", "c2", input, true)
	if len(interrupts(t, st)) != 2 || len(g.grants) != 0 {
		t.Fatalf("a mail delete was remembered: %v %v", interrupts(t, st), g.grants)
	}
}

func TestUnreadableInputAsksEveryTimeAndNeverGrants(t *testing.T) {
	g, st := rememberGate(t)
	input := `{"id":"abc"}`
	scopeAsk(t, g, "opus_update_task", "c1", input, true)
	scopeAsk(t, g, "opus_update_task", "c2", input, true)
	if len(interrupts(t, st)) != 2 {
		t.Fatalf("an unreadable call was granted instead of asked: %v", interrupts(t, st))
	}
	if len(g.grants) != 0 {
		t.Fatalf("an unreadable call recorded a grant: %v", g.grants)
	}
}

func TestSuitePolicyCardNamesResourceChangeAndSpace(t *testing.T) {
	g, _ := rememberGate(t)
	record, scope, ok := g.describe(context.Background(), "opus_update_task", json.RawMessage(`{"id":1,"title":"Ship it"}`))
	if !ok {
		t.Fatal("the card refused a readable call")
	}
	if scope != "task:1" {
		t.Fatalf("scope = %q, want task:1", scope)
	}
	fields, ok := record.(map[string]any)
	if !ok {
		t.Fatalf("record = %#v", record)
	}
	if fields["resource"] != "task" || fields["action"] != "update" || fields["space"] != "Work" {
		t.Fatalf("card record = %v", fields)
	}
	if fields["title"] != "Ship it" {
		t.Fatalf("the card hides the change: %v", fields)
	}
	if _, dup := fields["id"]; dup {
		t.Fatalf("the card repeats the bare id: %v", fields)
	}
	unreadable, unreadableScope, ok := g.describe(context.Background(), "opus_update_task", json.RawMessage(`not json`))
	if !ok {
		t.Fatal("an unreadable call must still reach the owner")
	}
	if unreadableScope != "" {
		t.Fatalf("an unreadable call got scope %q", unreadableScope)
	}
	if fields, ok := unreadable.(map[string]any); !ok || fields["tool"] != "opus_update_task" {
		t.Fatalf("unreadable record = %#v", unreadable)
	}
}

func TestApprovingOnceNeverRemembersASuiteWrite(t *testing.T) {
	g, st := rememberGate(t)
	scopeAsk(t, g, "opus_update_task", "c1", `{"id":1}`, false)
	scopeAsk(t, g, "opus_update_task", "c2", `{"id":1,"title":"again"}`, false)
	scopeAsk(t, g, "opus_create_task", "c3", `{"title":"a"}`, false)
	scopeAsk(t, g, "opus_create_task", "c4", `{"title":"b"}`, false)
	if len(interrupts(t, st)) != 4 {
		t.Fatalf("an approval without stop asking covered a later call: %v", interrupts(t, st))
	}
}

func TestMailMoveNeverTakesAGrant(t *testing.T) {
	g, st := rememberGate(t)
	g.outward["courrier_move"] = true
	defaults := suite.Defaults{CourrierAccount: 5}
	g.addPolicy("courrier_move", suitePolicy{name: "courrier_move", defaults: defaults, gate: g})
	scopeAsk(t, g, "courrier_move", "c1", `{"ids":[1],"folder":"Trash"}`, true)
	scopeAsk(t, g, "courrier_move", "c2", `{"ids":[2,3],"folder":"Trash"}`, true)
	if len(interrupts(t, st)) != 2 {
		t.Fatalf("a mail move was remembered: %v", interrupts(t, st))
	}
}
