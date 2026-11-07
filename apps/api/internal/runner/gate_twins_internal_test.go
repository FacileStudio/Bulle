package runner

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

type twin struct {
	approved chan int
	run      [2]chan struct{}
	out      [2]chan string
}

func startTwins(g *gate, tool nacelle.Tool, input string) *twin {
	tw := &twin{approved: make(chan int, 2)}
	for i := range 2 {
		tw.run[i], tw.out[i] = make(chan struct{}), make(chan string, 1)
		go func() {
			if !g.approve(context.Background(), mailSendTool, json.RawMessage(input)) {
				tw.out[i] <- "refused"
				return
			}
			tw.approved <- i
			<-tw.run[i]
			out, _ := tool.Run(context.Background(), json.RawMessage(input))
			tw.out[i] <- out
		}()
	}
	return tw
}

func waitPending(t *testing.T, g *gate, ids ...string) {
	t.Helper()
	for _, id := range ids {
		for deadline := time.Now().Add(2 * time.Second); !g.pending(id); {
			if time.Now().After(deadline) {
				t.Fatalf("%s never asked the owner", id)
			}
			time.Sleep(5 * time.Millisecond)
		}
	}
}

func TestTwinCallsGetTheirOwnCardAndEdit(t *testing.T) {
	g, tools := mailGate(t)
	input := `{"to":["a@x"],"body":"hi"}`
	g.note("c1", mailSendTool, input)
	g.note("c2", mailSendTool, input)
	tw := startTwins(g, tools[mailSendTool], input)
	waitPending(t, g, "c1", "c2")
	edit := Edit{Args: json.RawMessage(`{"body":"edited"}`)}
	if err := g.decide(context.Background(), "c1", true, "1", edit, false); err != nil {
		t.Fatal(err)
	}
	first := <-tw.approved
	if err := g.decide(context.Background(), "c2", true, "1", Edit{}, false); err != nil {
		t.Fatal(err)
	}
	select {
	case second := <-tw.approved:
		close(tw.run[second])
		close(tw.run[first])
	case <-time.After(200 * time.Millisecond):
		close(tw.run[first])
		close(tw.run[<-tw.approved])
	}
	edited, plain := <-tw.out[first], <-tw.out[1-first]
	if !strings.Contains(edited, `"body":"edited"`) || strings.Contains(plain, "edited") {
		t.Fatalf("the edit must run on the call it was made on: %q / %q", edited, plain)
	}
}

func TestClaimTakesTheExactCallFirst(t *testing.T) {
	g, _ := mailGate(t)
	g.note("c1", mailSendTool, `{"body":"one"}`)
	g.note("c2", mailSendTool, `{"body":"two"}`)
	if id := g.callID(context.Background(), mailSendTool, `{"body": "two"}`); id != "c2" {
		t.Fatalf("got %s, want the exact match c2", id)
	}
	if id := g.callID(context.Background(), mailSendTool, `{"body":"three"}`); id != "c1" {
		t.Fatalf("got %s, want the only queued call c1", id)
	}
	g.note("c3", mailSendTool, `{"body":"a"}`)
	g.note("c4", mailSendTool, `{"body":"b"}`)
	if id := g.callID(context.Background(), mailSendTool, `{"body":"z"}`); !strings.HasPrefix(id, "unannounced-") {
		t.Fatalf("an ambiguous call must not take a queued id, got %s", id)
	}
}

func TestCallIDWakesOnTheAnnouncement(t *testing.T) {
	g, _ := mailGate(t)
	got := make(chan string, 1)
	go func() { got <- g.callID(context.Background(), mailSendTool, `{"body":"late"}`) }()
	time.Sleep(20 * time.Millisecond)
	g.note("late", mailSendTool, `{"body":"late"}`)
	select {
	case id := <-got:
		if id != "late" {
			t.Fatalf("got %s", id)
		}
	case <-time.After(announceWait):
		t.Fatal("callID did not wake on note")
	}
}

type failingApprovals struct{ store.Store }

func (failingApprovals) CreateApproval(context.Context, string, string) error {
	return errors.New("database is gone")
}

func TestUnrecordedApprovalRefusesAndSaysSo(t *testing.T) {
	st := failingApprovals{store.NewMemory()}
	em := &emitter{store: st, hub: newHub(), logger: slog.Default(), runID: "r1"}
	g := newGate(map[string]bool{"send": true}, st, em, "r1", time.Now)
	g.note("c1", "send", `{}`)
	if g.approve(context.Background(), "send", json.RawMessage(`{}`)) {
		t.Fatal("a call whose approval was not recorded ran")
	}
	events, _ := st.Events(context.Background(), "r1", 0)
	if len(events) != 1 || events[0].Type != EventApprovalError || !strings.Contains(string(events[0].Payload), `"call_id":"c1"`) {
		t.Fatalf("want one APPROVAL_ERROR event, got %+v", events)
	}
}
