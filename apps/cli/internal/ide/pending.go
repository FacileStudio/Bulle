package ide

import (
	"fmt"

	"github.com/FacileStudio/nacelle"

	"github.com/FacileStudio/Bulle/apps/cli/internal/diff"
)

// callID is the id a call's start, finish and edit share: the one the model
// gave it, or its tool and position in the turn when a backend gave none.
func callID(call nacelle.Call) string {
	if call.ID != "" {
		return call.ID
	}
	return fmt.Sprintf("%s-%d", call.Name, call.Index)
}

// remember records the change a call is about to make.
func (p *publisher) remember(id string, change diff.EditChange) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pending[id] = change
}

// take returns the change a call made, and forgets it.
func (p *publisher) take(id string) (diff.EditChange, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	change, ok := p.pending[id]
	delete(p.pending, id)
	return change, ok
}
