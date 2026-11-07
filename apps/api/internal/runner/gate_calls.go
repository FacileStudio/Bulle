package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// announceWait bounds how long an asked call waits for the stream to name its
// call id before it gets an id of its own.
const announceWait = 500 * time.Millisecond

func normalize(input string) string {
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(input)); err != nil {
		return input
	}
	return buf.String()
}

// note records a call the stream announced, when that call will ask the
// owner, and wakes whoever waits for an id of that tool.
func (g *gate) note(id, name, input string) {
	if !g.asks(name, json.RawMessage(input)) {
		return
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.asked[name] = append(g.asked[name], announced{id: id, input: normalize(input)})
	if signal, ok := g.signals[name]; ok {
		close(signal)
		delete(g.signals, name)
	}
}

// claimLocked takes the announced call whose input matches exactly. Only when
// a single call of the tool is queued does it take that one whatever its
// input, since there is nothing else it could be. The caller holds g.mu.
func (g *gate) claimLocked(name, input string) (string, bool) {
	queue := g.asked[name]
	want := normalize(input)
	pick := -1
	for i, candidate := range queue {
		if candidate.input == want {
			pick = i
			break
		}
	}
	if pick < 0 && len(queue) == 1 {
		pick = 0
	}
	if pick < 0 {
		return "", false
	}
	id := queue[pick].id
	g.asked[name] = append(queue[:pick:pick], queue[pick+1:]...)
	return id, true
}

// callID is the id the stream gave this call. It waits for the announcement,
// bounded by ctx and announceWait, then falls back to an id of its own.
func (g *gate) callID(ctx context.Context, name, input string) string {
	timeout := time.NewTimer(announceWait)
	defer timeout.Stop()
	for {
		g.mu.Lock()
		if id, ok := g.claimLocked(name, input); ok {
			g.mu.Unlock()
			return id
		}
		signal, ok := g.signals[name]
		if !ok {
			signal = make(chan struct{})
			g.signals[name] = signal
		}
		g.mu.Unlock()
		select {
		case <-signal:
		case <-ctx.Done():
			return g.orphan()
		case <-timeout.C:
			return g.orphan()
		}
	}
}

func (g *gate) orphan() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.orphans++
	return fmt.Sprintf("unannounced-%d", g.orphans)
}
