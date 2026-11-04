package hardware_test

import (
	"context"
	"sync"
	"time"

	internal "github.com/FacileStudio/Bulle/apps/api/internal/hardware"
)

type fakeStore struct {
	mu      sync.Mutex
	devices map[string]internal.Device
	hashes  map[string]string
}

func newFakeStore() *fakeStore {
	return &fakeStore{devices: map[string]internal.Device{}, hashes: map[string]string{}}
}

func (f *fakeStore) Claim(_ context.Context, d internal.Device, hash string) (internal.Device, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	d.Capabilities = "[]"
	f.devices[d.ID] = d
	f.hashes[hash] = d.ID
	return d, nil
}

func (f *fakeStore) Devices(_ context.Context, userID int64) ([]internal.Device, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []internal.Device
	for _, d := range f.devices {
		if d.UserID == userID {
			out = append(out, d)
		}
	}
	return out, nil
}

func (f *fakeStore) Revoke(ctx context.Context, userID int64, id string) error {
	if _, err := f.Owned(ctx, userID, id); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.devices, id)
	return nil
}

func (f *fakeStore) Owned(_ context.Context, userID int64, id string) (internal.Device, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	d, ok := f.devices[id]
	if !ok || d.UserID != userID {
		return internal.Device{}, internal.ErrNotFound
	}
	return d, nil
}

func (f *fakeStore) ByTokenHash(_ context.Context, hash string) (internal.Device, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	d, ok := f.devices[f.hashes[hash]]
	if !ok {
		return internal.Device{}, internal.ErrNotFound
	}
	return d, nil
}

func (f *fakeStore) Touch(context.Context, string, string, time.Time) error { return nil }
