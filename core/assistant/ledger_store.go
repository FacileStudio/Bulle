package assistant

import (
	"encoding/json"
	"sync"
	"time"
)

// Entry is what the ledger remembers about one item key.
type Entry struct {
	FirstSeen    time.Time `json:"first_seen"`
	LastSeen     time.Time `json:"last_seen"`
	LastSurfaced time.Time `json:"last_surfaced,omitzero"`
	Fingerprint  string    `json:"fingerprint,omitempty"`
	SnoozedUntil time.Time `json:"snoozed_until,omitzero"`
	Noise        bool      `json:"noise,omitempty"`
	LastStatus   string    `json:"last_status,omitempty"`
	Phase        string    `json:"phase,omitempty"`
}

// Ledger records what the assistant already told the user so the same
// unchanged item is never surfaced twice. It is safe for concurrent use.
type Ledger struct {
	mu      sync.Mutex
	state   State
	entries map[string]*Entry
}

// NewLedger opens the ledger kept in st. A missing document starts empty. A
// corrupt one is copied to the "ledger.corrupt" document and the ledger
// starts empty.
func NewLedger(st State) (*Ledger, error) {
	l := &Ledger{state: st, entries: map[string]*Entry{}}
	data, err := st.Load("ledger")
	if err != nil {
		return nil, err
	}
	return l, l.decode(data)
}

func (l *Ledger) decode(data []byte) error {
	l.entries = map[string]*Entry{}
	if data == nil {
		return nil
	}
	if json.Unmarshal(data, &l.entries) != nil || l.entries == nil {
		l.entries = map[string]*Entry{}
		return l.state.Update("ledger.corrupt", func([]byte) ([]byte, error) { return data, nil })
	}
	return nil
}

func (l *Ledger) update(fn func()) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.state.Update("ledger", func(current []byte) ([]byte, error) {
		if err := l.decode(current); err != nil {
			return nil, err
		}
		fn()
		return json.MarshalIndent(l.entries, "", "  ")
	})
}

func (l *Ledger) refresh() {
	data, err := l.state.Load("ledger")
	if err != nil {
		return
	}
	if err := l.decode(data); err != nil {
		return
	}
}
