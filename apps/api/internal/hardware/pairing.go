package hardware

import (
	"errors"
	"sync"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

// PairingTTL is how long a pairing code stays valid.
const PairingTTL = 10 * time.Minute

// MaxPendingPairings caps the codes waiting at once across the server.
const MaxPendingPairings = 256

// ErrPairingBusy reports that too many codes are waiting.
var ErrPairingBusy = errors.New("hardware: too many pairings in progress")

// Pairing is a device that asked for a code and is waiting for a user to
// enter it.
type Pairing struct {
	HardwareID  string
	DeviceClass string
	Name        string
	UserCode    string
	DeviceCode  string
	ExpiresAt   time.Time
}

// Grant is what a paired device receives on its next poll, once.
type Grant struct {
	Device       Device
	DeviceToken  string
	GatewayWSURL string
}

type pending struct {
	Pairing
	grant *Grant
}

// Pairings keeps the codes of devices pairing without a BLE link, in memory:
// they live minutes, and a restart only makes the device ask again.
type Pairings struct {
	mu       sync.Mutex
	byUser   map[string]*pending
	byDevice map[string]*pending
	now      func() time.Time
}

// NewPairings answers an empty code store.
func NewPairings(now func() time.Time) *Pairings {
	if now == nil {
		now = time.Now
	}
	return &Pairings{byUser: map[string]*pending{}, byDevice: map[string]*pending{}, now: now}
}

// Start records a device waiting to be paired and answers its codes: the
// user code it shows, and the device code it polls with.
func (p *Pairings) Start(hardwareID, deviceClass, name string) (Pairing, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.prune()
	if len(p.byDevice) >= MaxPendingPairings {
		return Pairing{}, ErrPairingBusy
	}
	code := newUserCode()
	for p.byUser[code] != nil {
		code = newUserCode()
	}
	entry := &pending{Pairing: Pairing{
		HardwareID:  hardwareID,
		DeviceClass: deviceClass,
		Name:        name,
		UserCode:    code,
		DeviceCode:  store.NewID(),
		ExpiresAt:   p.now().Add(PairingTTL),
	}}
	p.byUser[code] = entry
	p.byDevice[entry.DeviceCode] = entry
	return entry.Pairing, nil
}

// Take removes a user code so no one else can enter it, and answers the
// device waiting behind it.
func (p *Pairings) Take(userCode string) (Pairing, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.prune()
	entry, ok := p.byUser[NormalizeCode(userCode)]
	if !ok {
		return Pairing{}, false
	}
	delete(p.byUser, entry.UserCode)
	return entry.Pairing, true
}

// Restore puts back a code taken by a claim that then failed.
func (p *Pairings) Restore(pairing Pairing) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if entry, ok := p.byDevice[pairing.DeviceCode]; ok && entry.grant == nil {
		p.byUser[pairing.UserCode] = entry
	}
}

// Complete hands the grant to the device behind pairing.
func (p *Pairings) Complete(pairing Pairing, grant Grant) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if entry, ok := p.byDevice[pairing.DeviceCode]; ok {
		entry.grant = &grant
	}
}

// Poll answers the grant of a device code once it is claimed. found is false
// for an unknown or expired code; a found code with no grant is still waiting.
func (p *Pairings) Poll(deviceCode string) (grant Grant, found, claimed bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.prune()
	entry, ok := p.byDevice[deviceCode]
	if !ok {
		return Grant{}, false, false
	}
	if entry.grant == nil {
		return Grant{}, true, false
	}
	delete(p.byDevice, deviceCode)
	return *entry.grant, true, true
}

func (p *Pairings) prune() {
	now := p.now()
	for code, entry := range p.byDevice {
		if now.After(entry.ExpiresAt) {
			delete(p.byDevice, code)
			delete(p.byUser, entry.UserCode)
		}
	}
}
