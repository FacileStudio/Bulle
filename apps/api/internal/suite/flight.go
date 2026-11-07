package suite

import "sync"

type flightCall struct {
	done  chan struct{}
	value string
	err   error
}

type flightGroup struct {
	mu    sync.Mutex
	calls map[string]*flightCall
}

func (g *flightGroup) do(key string, fn func() (string, error)) (string, error) {
	g.mu.Lock()
	if g.calls == nil {
		g.calls = map[string]*flightCall{}
	}
	if call, ok := g.calls[key]; ok {
		g.mu.Unlock()
		<-call.done
		return call.value, call.err
	}
	call := &flightCall{done: make(chan struct{})}
	g.calls[key] = call
	g.mu.Unlock()

	call.value, call.err = fn()
	g.mu.Lock()
	delete(g.calls, key)
	g.mu.Unlock()
	close(call.done)
	return call.value, call.err
}
